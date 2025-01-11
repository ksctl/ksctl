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

package addons

type ClusterAddon struct {
	Label string `json:"label"`
	IsCNI bool   `json:"cni,omitempty"`
	Name  string `json:"name"`

	// Config is a string representation of a JSON object
	Config *string `json:"config,omitempty"`
}

type ClusterAddons []ClusterAddon

func (ca ClusterAddons) GetAddons(label string) []ClusterAddon {
	var caa []ClusterAddon
	if len(ca) == 0 {
		return nil
	}
	for _, c := range ca {
		if c.Label == label {
			caa = append(caa, c)
		}
	}
	return caa
}

func (ca ClusterAddons) GetAddonLabels() []string {
	var labels []string
	if len(ca) == 0 {
		return nil
	}
	for _, c := range ca {
		labels = append(labels, c.Label)
	}
	return labels
}
