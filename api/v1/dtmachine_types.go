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

	DtNode   string `json:"dtnode,omitempty"`
	Type     string `json:"type,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	Cpu      int32  `json:"cpu,omitempty"`
	Memory   int64  `json:"memory,omitempty"`
	Disk     string `json:"disk,omitempty"`
	Base     string `json:"base,omitempty"`
	Os       string `json:"os,omitempty"`
	//负责部门
	//负责人
	//释放策略

}

// 状态信息
// MachineStatus defines the observed state of Machine
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
}

// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Ip",type=string,JSONPath=`.status.ip`
// +kubebuilder:printcolumn:name="DtNode",type=string,JSONPath=`.spec.dtnode`
// +kubebuilder:printcolumn:name="type",type=string,JSONPath=`.spec.type`
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Machine is the Schema for the machines API
type DtMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DtMachineSpec   `json:"spec,omitempty"`
	Status DtMachineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MachineList contains a list of Machine
type DtMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DtMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DtMachine{}, &DtMachineList{})
}
