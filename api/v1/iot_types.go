package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IotSpec defines the desired state of Iot
type IotSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Ip   string `json:"ip,omitempty"`
	Mac  string `json:"mac,omitempty"`
	Dns  string `json:"dns,omitempty"`
	Type string `json:"type,omitempty"`
}

// IotStatus defines the observed state of Iot
type IotStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Ready    string `json:"ready,omitempty"`
	NotReady string `json:"notready,omitempty"`
	Online   string `json:"online,omitempty"`
	Offline  string `json:"offlinema,omitempty"`
}

// +kubebuilder:printcolumn:name="Ip",type=string,JSONPath=`.spec.ip`
// +kubebuilder:printcolumn:name="Mac",type=string,JSONPath=`.spec.mac`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Dns",type=string,JSONPath=`.spec.dns`
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Iot is the Schema for the iots API
type Iot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IotSpec   `json:"spec,omitempty"`
	Status IotStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IotList contains a list of Iot
type IotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Iot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Iot{}, &IotList{})
}
