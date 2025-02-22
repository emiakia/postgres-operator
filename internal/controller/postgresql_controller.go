package controller

import (
    "context"

    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/types"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"

    databasev1alpha1 "github.com/example/postgresql-operator/api/v1alpha1"
)

type PostgreSQLReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=database.example.com,resources=postgresqls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=database.example.com,resources=postgresqls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=database.example.com,resources=postgresqls/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *PostgreSQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)
    logger.Info("Reconciling PostgreSQL", "namespace", req.Namespace, "name", req.Name)

    postgresql := &databasev1alpha1.PostgreSQL{}
    err := r.Get(ctx, req.NamespacedName, postgresql)
    if err != nil {
        if errors.IsNotFound(err) {
            logger.Info("PostgreSQL resource not found. Ignoring since object must be deleted")
            return ctrl.Result{}, nil
        }
        logger.Error(err, "Failed to get PostgreSQL")
        return ctrl.Result{}, err
    }

    // Create Service
    if err := r.reconcileService(ctx, postgresql); err != nil {
        logger.Error(err, "Failed to reconcile Service")
        return ctrl.Result{}, err
    }

    // Create StatefulSet
    if err := r.reconcileStatefulSet(ctx, postgresql); err != nil {
        logger.Error(err, "Failed to reconcile StatefulSet")
        return ctrl.Result{}, err
    }

    logger.Info("Reconciliation completed successfully")
    return ctrl.Result{}, nil
}

func (r *PostgreSQLReconciler) reconcileService(ctx context.Context, postgresql *databasev1alpha1.PostgreSQL) error {
    logger := log.FromContext(ctx)

    svc := &corev1.Service{}
    err := r.Get(ctx, types.NamespacedName{Name: postgresql.Name, Namespace: postgresql.Namespace}, svc)
    if err != nil && errors.IsNotFound(err) {
        // Define new service
        svc = &corev1.Service{
            ObjectMeta: metav1.ObjectMeta{
                Name:      postgresql.Name,
                Namespace: postgresql.Namespace,
            },
            Spec: corev1.ServiceSpec{
                Selector: map[string]string{
                    "app": postgresql.Name,
                },
                Ports: []corev1.ServicePort{{
                    Port:     5432,
                    Protocol: corev1.ProtocolTCP,
                    Name:     "postgresql",
                }},
            },
        }
        logger.Info("Creating new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
        return r.Create(ctx, svc)
    }
    return err
}

func (r *PostgreSQLReconciler) reconcileStatefulSet(ctx context.Context, postgresql *databasev1alpha1.PostgreSQL) error {
    logger := log.FromContext(ctx)

    sts := &appsv1.StatefulSet{}
    err := r.Get(ctx, types.NamespacedName{Name: postgresql.Name, Namespace: postgresql.Namespace}, sts)
    if err != nil && errors.IsNotFound(err) {
        // Define new statefulset
        sts = &appsv1.StatefulSet{
            ObjectMeta: metav1.ObjectMeta{
                Name:      postgresql.Name,
                Namespace: postgresql.Namespace,
            },
            Spec: appsv1.StatefulSetSpec{
                ServiceName: postgresql.Name,
                Replicas:   &postgresql.Spec.Size,
                Selector: &metav1.LabelSelector{
                    MatchLabels: map[string]string{
                        "app": postgresql.Name,
                    },
                },
                Template: corev1.PodTemplateSpec{
                    ObjectMeta: metav1.ObjectMeta{
                        Labels: map[string]string{
                            "app": postgresql.Name,
                        },
                    },
                    Spec: corev1.PodSpec{
                        Containers: []corev1.Container{{
                            Name:  "postgresql",
                            Image: postgresql.Spec.Image,
                            Env: []corev1.EnvVar{
                                {
                                    Name:  "POSTGRES_DB",
                                    Value: postgresql.Spec.DatabaseName,
                                },
                                {
                                    Name:  "POSTGRES_USER",
                                    Value: postgresql.Spec.Username,
                                },
                                {
                                    Name:  "POSTGRES_PASSWORD",
                                    Value: postgresql.Spec.Password,
                                },
                                {
                                    Name:  "PGDATA",
                                    Value: "/var/lib/postgresql/data/pgdata",
                                },
                            },
                            Ports: []corev1.ContainerPort{{
                                ContainerPort: 5432,
                                Name:         "postgresql",
                            }},
                            VolumeMounts: []corev1.VolumeMount{{
                                Name:      "data",
                                MountPath: "/var/lib/postgresql/data",
                            }},
                        }},
                    },
                },
                VolumeClaimTemplates: []corev1.PersistentVolumeClaim{{
                    ObjectMeta: metav1.ObjectMeta{
                        Name: "data",
                    },
                    Spec: corev1.PersistentVolumeClaimSpec{
                        AccessModes: []corev1.PersistentVolumeAccessMode{
                            corev1.ReadWriteOnce,
                        },
                        Resources: corev1.VolumeResourceRequirements{
                            Requests: corev1.ResourceList{
                                corev1.ResourceStorage: resource.MustParse(postgresql.Spec.StorageSize),
                            },
                        },
                    },
                }},
            },
        }
        logger.Info("Creating new StatefulSet", "StatefulSet.Namespace", sts.Namespace, "StatefulSet.Name", sts.Name)
        return r.Create(ctx, sts)
    }
    return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgreSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&databasev1alpha1.PostgreSQL{}).
        Owns(&appsv1.StatefulSet{}).
        Owns(&corev1.Service{}).
        Complete(r)
}
