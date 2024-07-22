/*
Copyright 2024.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:pruning:PreserveUnknownFields
// +kubebuilder:validation:Schemaless
type ComponentOverrides unstructured.Unstructured

func (in *ComponentOverrides) DeepCopyInto(out *ComponentOverrides) {
	if out != nil {
		casted := unstructured.Unstructured(*in)
		deepCopy := casted.DeepCopy()
		out.Object = deepCopy.Object
	}
}

type ComponentId string

// StackSpec defines the desired state of Stack
type StackSpec struct {
	StackName string `json:"stackName"`

	//+kubebuilder:pruning:PreserveUnknownFields
	//+kubebuilder:validation:Schemaless
	//+optional
	Overrides map[ComponentId]ComponentOverrides `json:"overrides"`
}

// StackStatus defines the observed state of Stack
type StackStatus struct {
	Success         bool   `json:"success"`
	ReasonOfFailure string `json:"reasonOfFailure,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Stack is the Schema for the stacks API
type Stack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StackSpec   `json:"spec,omitempty"`
	Status StackStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StackList contains a list of Stack
type StackList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Stack `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Stack{}, &StackList{})
}
