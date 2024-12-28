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

package handler

import (
	"strings"

	"github.com/ksctl/ksctl/pkg/helm"
	"github.com/ksctl/ksctl/pkg/poller"
	"github.com/ksctl/ksctl/pkg/utilities"
)

func getCiliumComponentOverridings(p ComponentOverrides) (version *string, ciliumChartOverridings map[string]any) {
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

func setCiliumComponentOverridings(p ComponentOverrides) (
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
		ciliumChartOverridings = nil
	}
	return
}

func CiliumStandardComponent(params ComponentOverrides) (StackComponent, error) {
	version, ciliumChartOverridings, err := setCiliumComponentOverridings(params)
	if err != nil {
		return StackComponent{}, err
	}

	if strings.HasPrefix(version, "v") {
		version = strings.TrimPrefix(version, "v")
	}

	return StackComponent{
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
		HandlerType: ComponentTypeHelm,
	}, nil
}

func CiliumStandardCNI(params ApplicationParams) (ApplicationStack, error) {
	v, err := CiliumStandardComponent(
		params.ComponentParams[CiliumComponentID],
	)
	if err != nil {
		return ApplicationStack{}, err
	}

	return ApplicationStack{
		Components: map[StackComponentID]StackComponent{
			CiliumComponentID: v,
		},

		StkDepsIdx:  []StackComponentID{CiliumComponentID},
		StackNameID: CiliumStandardStackID,
		Maintainer:  "github@dipankardas011",
	}, nil
}
