package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Machine结构体
// DtMachineSpec defines the desired state of Machine
type DtMachineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Labels         map[string]string `json:"labels,omitempty"`
	Desc           string            `json:"desc,omitempty"`
	Dept           string            `json:"dept,omitempty"`
	Mantainer      string            `json:"mantainer,omitempty"`
	ReleaseStragle string            `json:"releasestragle,omitempty"`
	ReleaseDate    string            `json:"releasedate,omitempty"`

	DtCluster string `json:"dtcluster,omitempty"`
	DtModel   string `json:"dtmodel,omitempty"`
}

// 状态信息
// MachineStatus defines the observed state of DtMachine
type DtMachineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase      string `json:"phase,omitempty"`
	HostName   string `json:"hostname,omitempty"`
	Ip         string `json:"ip,omitempty"`
	Mac        string `json:"mac,omitempty"`
	CpuUsed    string `json:"cpuused,omitempty"`
	MemoryUsed string `json:"memoryused,omitempty"`
	DiskUsed   string `json:"diskused,omitempty"`

	DtNode string `json:"dtnode,omitempty"`
	Os     string `json:"os,omitempty"`
}

//+kubebuilder:printcolumn:name="DtCluster",type=string,JSONPath=`.spec.dtcluster`
//+kubebuilder:printcolumn:name="DtNode",type=string,JSONPath=`.status.dtnode`
//+kubebuilder:printcolumn:name="DtModel",type=string,JSONPath=`.spec.dtmodel`
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="Ip",type=string,JSONPath=`.status.ip`
//+kubebuilder:printcolumn:name="Os",type=string,JSONPath=`.status.os`
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DtMachine is the Schema for the dtmachines API
type DtMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DtMachineSpec   `json:"spec,omitempty"`
	Status DtMachineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MachineList contains a list of DtMachine
type DtMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DtMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DtMachine{}, &DtMachineList{})
}
