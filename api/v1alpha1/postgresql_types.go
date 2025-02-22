package v1alpha1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PostgreSQLSpec defines the desired state of PostgreSQL
type PostgreSQLSpec struct {
    // Size is the number of PostgreSQL instances
    Size int32 `json:"size"`

    // StorageSize is the size of persistent volume
    StorageSize string `json:"storageSize"`

    // Image is the PostgreSQL docker image
    Image string `json:"image"`

    // DatabaseName is the name of the database to create
    DatabaseName string `json:"databaseName"`

    // Username is the admin username
    Username string `json:"username"`

    // Password is the admin password
    Password string `json:"password"`
}

// PostgreSQLStatus defines the observed state of PostgreSQL
type PostgreSQLStatus struct {
    // Nodes are the names of the PostgreSQL pods
    Nodes []string `json:"nodes"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PostgreSQL is the Schema for the postgresqls API
type PostgreSQL struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   PostgreSQLSpec   `json:"spec,omitempty"`
    Status PostgreSQLStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PostgreSQLList contains a list of PostgreSQL
type PostgreSQLList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []PostgreSQL `json:"items"`
}

func init() {
    SchemeBuilder.Register(&PostgreSQL{}, &PostgreSQLList{})
}
