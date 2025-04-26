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

package cni

import (
	"strings"

	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
	"github.com/ksctl/ksctl/v2/pkg/helm"
	"github.com/ksctl/ksctl/v2/pkg/poller"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

const (
	CiliumGuidedHubble     = "hubble"
	CiliumGuidedEncryption = "encryption"
	CiliumGuidedPrometheus = "prometheus"
)

type CiliumGuidedOutput struct {
	Name        string
	Description string
}

func CiliumGuidedConfigurations() []CiliumGuidedOutput {
	return []CiliumGuidedOutput{
		{
			Name:        CiliumGuidedHubble,
			Description: "Enable Hubble UI and Relay",
		},
		{
			Name:        CiliumGuidedEncryption,
			Description: "Enable Wireguard encryption",
		},
		{
			Name:        CiliumGuidedPrometheus,
			Description: "Enable Prometheus metrics",
		},
	}
}

func getCiliumComponentOverridings(p stack.ComponentOverrides) (version *string, ciliumChartOverridings map[string]any, guidedConfig []string) {
	ciliumChartOverridings = nil // By default it is nil

	if p == nil {
		return nil, nil, nil
	}

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "guidedConfig":
			if v, ok := v.([]string); ok {
				guidedConfig = v
			}

		case "ciliumChartOverridings":
			// https://artifacthub.io/packages/helm/cilium/cilium?modal=values-schema
			if v, ok := v.(map[string]any); ok {
				ciliumChartOverridings = v
			}
		}
	}
	return
}

func setCiliumComponentOverridings(p stack.ComponentOverrides) (
	version string,
	ciliumChartOverridings map[string]any,
	err error,
) {
	releases, err := poller.GetSharedPoller().Get("cilium", "cilium")
	if err != nil {
		return "", nil, err
	}

	ciliumChartOverridings = map[string]any{}

	_version, _ciliumChartOverridings, _guidedSetup := getCiliumComponentOverridings(p)

	version = getVersionIfItsNotNilAndLatest(_version, releases[0])

	if _ciliumChartOverridings != nil {
		ciliumChartOverridings = _ciliumChartOverridings
	} else if _guidedSetup != nil {
		for _, v := range _guidedSetup {
			if v == CiliumGuidedHubble {
				ciliumChartOverridings["hubble"] = map[string]any{
					"ui":    map[string]any{"enabled": true},
					"relay": map[string]any{"enabled": true},
					"metrics": map[string]any{"enabled": []string{
						"dns",
						"drop",
						"tcp",
						"flow",
						"port-distribution",
						"icmp",
						"httpV2:exemplars=true;labelsContext=source_ip,source_namespace,source_workload,destination_ip,destination_namespace,destination_workload,traffic_direction",
					}},
				}
			} else if v == CiliumGuidedEncryption {
				ciliumChartOverridings["encryption"] = map[string]any{
					"enabled": true,
					"type":    "wireguard",
				}
			} else if v == CiliumGuidedPrometheus {
				ciliumChartOverridings["operator"] = map[string]any{
					"replicas": 3,
					"prometheus": map[string]any{
						"enabled": true,
					},
				}
				ciliumChartOverridings["prometheus"] = map[string]any{
					"enabled": true,
				}
			}
		}

	} else {
		ciliumChartOverridings = map[string]any{
			"hubble": map[string]any{
				"ui":    map[string]any{"enabled": true},
				"relay": map[string]any{"enabled": true},
				"metrics": map[string]any{"enabled": []string{
					"dns",
					"drop",
					"tcp",
					"flow",
					"port-distribution",
					"icmp",
					"httpV2:exemplars=true;labelsContext=source_ip,source_namespace,source_workload,destination_ip,destination_namespace,destination_workload,traffic_direction",
				}},
			},
			"encryption": map[string]any{
				"enabled": true,
				"type":    "wireguard",
			},
			"operator": map[string]any{
				"replicas": 3,
				"prometheus": map[string]any{
					"enabled": true,
				},
			},
			"prometheus": map[string]any{
				"enabled": true,
			},
		}
	}

	return
}

func CiliumStandardComponent(params stack.ComponentOverrides) (stack.Component, error) {
	version, ciliumChartOverridings, err := setCiliumComponentOverridings(params)
	if err != nil {
		return stack.Component{}, err
	}

	if strings.HasPrefix(version, "v") {
		version = strings.TrimPrefix(version, "v")
	}

	return stack.Component{
		Helm: &helm.App{
			RepoName: "cilium",
			RepoUrl:  "https://helm.cilium.io/",
			Charts: []helm.ChartOptions{
				{
					Name:            "cilium/cilium",
					Version:         version,
					ReleaseName:     "cilium",
					Namespace:       "kube-system",
					CreateNamespace: false,
					Args:            ciliumChartOverridings,
				},
			},
		},
		HandlerType: stack.ComponentTypeHelm,
	}, nil
}

func CiliumStandardCNI(params stack.ApplicationParams) (stack.ApplicationStack, error) {
	v, err := CiliumStandardComponent(
		params.ComponentParams[CiliumComponentID],
	)
	if err != nil {
		return stack.ApplicationStack{}, err
	}

	return stack.ApplicationStack{
		Components: map[stack.ComponentID]stack.Component{
			CiliumComponentID: v,
		},

		StkDepsIdx:  []stack.ComponentID{CiliumComponentID},
		StackNameID: CiliumStandardStackID,
		Maintainer:  "github@dipankardas011",
	}, nil
}
