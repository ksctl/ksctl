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
	"fmt"

	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
	"github.com/ksctl/ksctl/v2/pkg/k8s"
	"github.com/ksctl/ksctl/v2/pkg/poller"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

func getFlannelComponentOverridings(p stack.ComponentOverrides) (version *string) {
	if p == nil {
		return nil
	}

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		}
	}
	return
}

func setFlannelComponentOverridings(p stack.ComponentOverrides) (
	version string,
	url string,
	postInstall string,
	err error,
) {
	releases, err := poller.GetSharedPoller().Get("flannel-io", "flannel")
	if err != nil {
		return "", "", "", err
	}

	url = ""
	postInstall = ""

	_version := getFlannelComponentOverridings(p)
	version = getVersionIfItsNotNilAndLatest(_version, releases[0])

	defaultVals := func() {
		url = fmt.Sprintf("https://github.com/flannel-io/flannel/releases/download/%s/kube-flannel.yml", version)
		postInstall = "https://github.com/flannel-io/flannel"
	}

	defaultVals()
	return
}

func FlannelStandardComponent(params stack.ComponentOverrides) (stack.Component, error) {
	version, url, postInstall, err := setFlannelComponentOverridings(params)
	if err != nil {
		return stack.Component{}, err
	}

	return stack.Component{
		HandlerType: stack.ComponentTypeKubectl,
		Kubectl: &k8s.App{
			Urls:            []string{url},
			Version:         version,
			CreateNamespace: false,
			Metadata:        fmt.Sprintf("Flannel (Ver: %s) is a simple and easy way to configure a layer 3 network fabric designed for Kubernetes.", version),
			PostInstall:     postInstall,
		},
	}, nil
}

func FlannelStandardCNI(params stack.ApplicationParams) (stack.ApplicationStack, error) {
	v, err := FlannelStandardComponent(
		params.ComponentParams[FlannelComponentID],
	)
	if err != nil {
		return stack.ApplicationStack{}, err
	}

	return stack.ApplicationStack{
		Components: map[stack.ComponentID]stack.Component{
			FlannelComponentID: v,
		},
		StkDepsIdx:  []stack.ComponentID{FlannelComponentID},
		Maintainer:  "github:dipankardas011",
		StackNameID: FlannelStandardStackID,
	}, nil
}
