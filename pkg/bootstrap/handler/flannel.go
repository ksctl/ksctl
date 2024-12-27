package handler

import (
	"fmt"

	"github.com/ksctl/ksctl/pkg/k8s"
	"github.com/ksctl/ksctl/pkg/poller"
	"github.com/ksctl/ksctl/pkg/utilities"
)

func getFlannelComponentOverridings(p ComponentOverrides) (version *string) {
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

func setFlannelComponentOverridings(p ComponentOverrides) (
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

func FlannelStandardComponent(params ComponentOverrides) (StackComponent, error) {
	version, url, postInstall, err := setFlannelComponentOverridings(params)
	if err != nil {
		return StackComponent{}, err
	}

	return StackComponent{
		HandlerType: ComponentTypeKubectl,
		Kubectl: &k8s.App{
			Urls:            []string{url},
			Version:         version,
			CreateNamespace: false,
			Metadata:        fmt.Sprintf("Flannel (Ver: %s) is a simple and easy way to configure a layer 3 network fabric designed for Kubernetes.", version),
			PostInstall:     postInstall,
		},
	}, nil
}

func FlannelStandardCNI(params ApplicationParams) (ApplicationStack, error) {
	v, err := FlannelStandardComponent(
		params.ComponentParams[FlannelComponentID],
	)
	if err != nil {
		return ApplicationStack{}, err
	}

	return ApplicationStack{
		Components: map[StackComponentID]StackComponent{
			FlannelComponentID: v,
		},
		StkDepsIdx:  []StackComponentID{FlannelComponentID},
		Maintainer:  "github:dipankardas011",
		StackNameID: FlannelStandardStackID,
	}, nil
}
