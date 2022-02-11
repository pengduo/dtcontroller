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

// DtNode结构体
// DtNodeSpec defines the desired state of DtNode
type DtNodeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Labels map[string]string `json:"labels,omitempty"`

	Provider  string            `json:"provider,omitempty"`
	DtCluster map[string]string `json:"dtcluster,omitempty"`
	Desc      string            `json:"desc,omitempty"`
}

// 状态信息
// DtNodeStatus defines the observed state of DtNode
type DtNodeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase string `json:"phase,omitempty"`
	Node  string `json:"node,omitempty"`
}

//+kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.provider`
//+kubebuilder:printcolumn:name="Node",type=string,JSONPath=`.status.node`
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
