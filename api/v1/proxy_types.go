package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProxySpec defines the desired state of Proxy
type ProxySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Type      string `json:"type,omitempty"`
	ProxyName string `json:"proxyname,omitempty"`
}

// ProxyStatus defines the observed state of Proxy
type ProxyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Ready    string `json:"ready,omitempty"`
	NotReady string `json:"notready,omitempty"`
	Online   string `json:"online,omitempty"`
	Offline  string `json:"offlinema,omitempty"`
}

// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.Type`
// +kubebuilder:printcolumn:name="ProxyName",type=string,JSONPath=`.spec.proxyname`
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Proxy is the Schema for the proxies API
type Proxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProxySpec   `json:"spec,omitempty"`
	Status ProxyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProxyList contains a list of Proxy
type ProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Proxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Proxy{}, &ProxyList{})
}
