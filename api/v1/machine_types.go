package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MachineSpec defines the desired state of Machine
type MachineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	HostName  string `json:"hostname,omitempty"`
	Ip        string `json:"ip,omitempty"`
	Mac       string `json:"mac,omitempty"`
	User      string `json:"user,omitempty"`
	Password  string `json:"password,omitempty"`
	Cpu       string `json:"cpu,omitempty"`
	NodeName  string `json:"nodename,omitempty"`
	Command   string `json:"command,omitempty"`
	CmdResult string `json:"cmdresult,omitempty"`

	Labels map[string]string `json:"labels,omitempty"`
}

// MachineStatus defines the observed state of Machine
type MachineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase string `json:"phase,omitempty"`
}

// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="HostName",type=string,JSONPath=`.spec.hostname`
// +kubebuilder:printcolumn:name="Ip",type=string,JSONPath=`.spec.ip`
// +kubebuilder:printcolumn:name="Mac",type=string,JSONPath=`.spec.mac`
// +kubebuilder:printcolumn:name="User",type=string,JSONPath=`.spec.user`
// +kubebuilder:printcolumn:name="Cpu",type=string,JSONPath=`.spec.cpu`
// +kubebuilder:printcolumn:name="NodeName",type=string,JSONPath=`.spec.nodename`
// +kubebuilder:printcolumn:name="Labels",type=string,JSONPath=`.spec.labels`
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Machine is the Schema for the machines API
type Machine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineSpec   `json:"spec,omitempty"`
	Status MachineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MachineList contains a list of Machine
type MachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Machine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Machine{}, &MachineList{})
}
