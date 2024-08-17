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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ApplicationType string
type StackStatusCode string

const (
	TypeCNI ApplicationType = "cni"
	TypeApp ApplicationType = "app"

	Success StackStatusCode = "success"
	Failure StackStatusCode = "failure"
)

type StackObj struct {
	StackId string          `json:"stackId"`
	AppType ApplicationType `json:"appType"`

	Overrides *apiextensionsv1.JSON `json:"overrides,omitempty"`
}

// StackSpec defines the desired state of Stack
type StackSpec struct {
	Stacks []StackObj `json:"stacks"`
}

// StackStatus defines the observed state of Stack
type StackStatus struct {
	StatusCode      StackStatusCode `json:"statusCode,omitempty"`
	ReasonOfFailure string          `json:"reasonOfFailure,omitempty"`
}

// TODO: need to update the kind as `StackCollection`

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Stack is the Schema for the stacks API
type Stack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StackSpec   `json:"spec"`
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
