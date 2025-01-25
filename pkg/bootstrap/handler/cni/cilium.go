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

func getCiliumComponentOverridings(p stack.ComponentOverrides) (version *string, ciliumChartOverridings map[string]any) {
	ciliumChartOverridings = nil // By default it is nil

	if p == nil {
		return nil, nil
	}

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "ciliumChartOverridings":
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

	_version, _ciliumChartOverridings := getCiliumComponentOverridings(p)

	version = getVersionIfItsNotNilAndLatest(_version, releases[0])

	if _ciliumChartOverridings != nil {
		ciliumChartOverridings = _ciliumChartOverridings
	} else {
		ciliumChartOverridings = map[string]any{
			"hubble": map[string]any{
				"ui": map[string]any{
					"enabled": true,
				},
				"relay": map[string]any{
					"enabled": true,
				},
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
