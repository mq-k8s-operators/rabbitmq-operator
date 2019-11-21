package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RabbitmqSpec defines the desired state of Rabbitmq
// +k8s:openapi-gen=true
type RabbitmqSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Size               int32             `json:"size"`
	Image              string            `json:"image,omitempty"`
	NameSpace          string            `json:"namespace,omitempty"`
	Envs               []corev1.EnvVar   `json:"envs,omitempty"`
	Storage            resource.Quantity `json:"storage,omitempty"`
	Data               map[string]string `json:"data,omitempty"`
	Name               string            `json:"name,omitempty"`
	ServiceAccountName string            `json:"serviceAccountName,omitempty"`
	StorageClassName   string            `json:"storageClassName,omitempty"`
	PvLable            map[string]string `json:"pvLabel,omitempty"`
	Bash               []string          `json:"bash,omitempty"`
}

// RabbitmqStatus defines the observed state of Rabbitmq
// +k8s:openapi-gen=true
type RabbitmqStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Rabbitmq is the Schema for the rabbitmqs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=rabbitmqs,scope=Namespaced
type Rabbitmq struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitmqSpec   `json:"spec,omitempty"`
	Status RabbitmqStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RabbitmqList contains a list of Rabbitmq
type RabbitmqList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rabbitmq `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Rabbitmq{}, &RabbitmqList{})
}
