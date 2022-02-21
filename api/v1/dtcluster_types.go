package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DtClusterSpec defines the desired state of DtCluster
type DtClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Labels   map[string]string `json:"labels,omitempty"`
	Provider string            `json:"provider,omitempty"`
	Desc     string            `json:"desc,omitempty"`

	Content map[string]string `json:"content,omitempty"` //内容
}

// DtClusterStatus defines the observed state of DtCluster
type DtClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Bound  bool   `json:"bound,omitempty"`
	DtNode string `json:"dtnode,omitempty"`
}

//+kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.provider`
//+kubebuilder:printcolumn:name="Desc",type=string,JSONPath=`.spec.desc`
//+kubebuilder:printcolumn:name="Bound",type=string,JSONPath=`.status.bound`
//+kubebuilder:printcolumn:name="DtNode",type=string,JSONPath=`.status.dtnode`
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// DtCluster is the Schema for the dtclusters API
type DtCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DtClusterSpec   `json:"spec,omitempty"`
	Status DtClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DtClusterList contains a list of DtCluster
type DtClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DtCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DtCluster{}, &DtClusterList{})
}
