// Copyright 2025 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stack

import (
	"github.com/ksctl/ksctl/v2/pkg/helm"
	"github.com/ksctl/ksctl/v2/pkg/k8s"
)

type (
	ComponentType uint
	ComponentID   string
	ID            string
)

type Component struct {
	Helm        *helm.App
	Kubectl     *k8s.App
	HandlerType ComponentType
}

// TODO: need to think of taking some sport of the application ksctl provide from the src to some json file in ver control
//
//	so that we can update that and no need of update of the logicial part
//
// also add a String()

type ApplicationStack struct {
	Components map[ComponentID]Component

	// StkDepsIdx helps you to get sequence of components, aka it acts as a key value table
	StkDepsIdx []ComponentID

	Maintainer  string
	StackNameID ID
}

type ApplicationParams struct {
	// StkOverrides   map[string]any
	ComponentParams map[ComponentID]ComponentOverrides
}

type ComponentOverrides map[string]any

const (
	ComponentTypeHelm    ComponentType = iota
	ComponentTypeKubectl ComponentType = iota
)

func GetComponentVersionOverriding(component Component) string {
	if component.HandlerType == ComponentTypeKubectl {
		return component.Kubectl.Version
	}
	return component.Helm.Charts[0].Version
}

type KsctlApp struct {
	StackName string                    `json:"stack_name"`
	Overrides map[string]map[string]any `json:"overrides"`
}
