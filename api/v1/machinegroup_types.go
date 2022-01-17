package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MachineGroupSpec defines the desired state of MachineGroup
type MachineGroupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	DtNode   string `json:"dtnode,omitempty"`
	Type     string `json:"type,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	Cpu      int32  `json:"cpu,omitempty"`
	Memory   int64  `json:"memory,omitempty"`
	Disk     string `json:"disk,omitempty"`
	Base     string `json:"base,omitempty"`
	Os       string `json:"os,omitempty"`

	Rs int32 `json:"rs,omitempty"`
}

// MachineGroupStatus defines the observed state of MachineGroup
type MachineGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase string `json:"phase,omitempty"`

	Rs string `json:"rs,omitempty"`
}

//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="Rs",type=string,JSONPath=`.status.rs`
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MachineGroup is the Schema for the machinegroups API
type MachineGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineGroupSpec   `json:"spec,omitempty"`
	Status MachineGroupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MachineGroupList contains a list of MachineGroup
type MachineGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MachineGroup{}, &MachineGroupList{})
}
