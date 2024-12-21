// Copyright 2024 Ksctl Authors
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

package statefile

import (
	"fmt"
	"strings"
)

type KubernetesAddons struct {
	Apps []Application `json:"apps" bson:"apps"`
	Cni  Application   `json:"cni" bson:"cni"`
}

type Application struct {
	Name       string               `json:"name" bson:"name"`
	Components map[string]Component `json:"components" bson:"components"`
}

type Component struct {
	Version string `json:"version" bson:"version"`
}

func (c Component) String() string {
	return c.Version
}

func (a Application) String() string {
	if len(a.Name) == 0 {
		return ""
	}
	var components []string
	for _, c := range a.Components {
		x := ""
		if len(c.Version) == 0 {
			c.Version = "latest"
			// continue
		}
		x = c.String()
		components = append(components, x)
	}

	return fmt.Sprintf("%s::[%s]", a.Name, strings.Join(components, ","))
}
