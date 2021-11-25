/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DtNodeSpec defines the desired state of DtNode
type DtNodeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Provider string `json:"provider,omitempty"`
	Ip       string `json:"ip,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

// DtNodeStatus defines the observed state of DtNode
type DtNodeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase    string `json:"phase,omitempty"`
	Age      string `json:"age,omitempty"`
	Version  string `json:"version,omitempty"`
	Cpu      string `json:"cpu,omitempty"`
	Memory   string `json:"memory,omitempty"`
	Storage  string `json:"storage,omitempty"`
	Vms      int    `json:"vms,omitempty"`
	Hosts    int    `json:"hosts,omitempty"`
	Networks int    `json:"networks,omitempty"`
	TTL      string `json:"ttl,omitempty"`
}

//+kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.provider`
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.age`
//+kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.version`
//+kubebuilder:printcolumn:name="Cpu",type=string,JSONPath=`.status.cpu`
//+kubebuilder:printcolumn:name="Memory",type=string,JSONPath=`.status.memory`
//+kubebuilder:printcolumn:name="Storage",type=string,JSONPath=`.status.storage`
//+kubebuilder:printcolumn:name="Vms",type=integer,JSONPath=`.status.vms`
//+kubebuilder:printcolumn:name="Hosts",type=integer,JSONPath=`.status.hosts`
//+kubebuilder:printcolumn:name="Networks",type=integer,JSONPath=`.status.networks`
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// DtNode is the Schema for the dtnodes API
type DtNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DtNodeSpec   `json:"spec,omitempty"`
	Status DtNodeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DtNodeList contains a list of DtNode
type DtNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DtNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DtNode{}, &DtNodeList{})
}
