package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DtModelSpec defines the desired state of DtModel
type DtModelSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Labels   map[string]string `json:"labels,omitempty"`
	Desc     string            `json:"desc,omitempty"`
	Provider string            `json:"provider,omitempty"`
	Type     string            `json:"type,omitempty"` //类型
	Os       string            `json:"os,omitempty"`   //基本信息os
	Cpu      int32             `json:"cpu,omitempty"`
	Memory   int64             `json:"memory,omitempty"`
	Disk     int64             `json:"disk,omitempty"`

	Content map[string]string `json:"content,omitempty"` //内容
}

// DtModelStatus defines the observed state of DtModel
type DtModelStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase string `json:"phase,omitempty"`
	Bound bool   `json:"bound,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DtModel is the Schema for the dtmodels API
type DtModel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DtModelSpec   `json:"spec,omitempty"`
	Status DtModelStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DtModelList contains a list of DtModel
type DtModelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DtModel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DtModel{}, &DtModelList{})
}
