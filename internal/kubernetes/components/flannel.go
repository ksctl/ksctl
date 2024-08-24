package components

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/poller"
)

func getFlannelComponentOverridings(p metadata.ComponentOverrides) (version *string) {
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

func setFlannelComponentOverridings(p metadata.ComponentOverrides) (
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

func FlannelStandardComponent(params metadata.ComponentOverrides) (metadata.StackComponent, error) {
	version, url, postInstall, err := setFlannelComponentOverridings(params)
	if err != nil {
		return metadata.StackComponent{}, err
	}

	return metadata.StackComponent{
		HandlerType: metadata.ComponentTypeKubectl,
		Kubectl: &metadata.KubectlHandler{
			Url:             url,
			Version:         version,
			CreateNamespace: false,
			Metadata:        fmt.Sprintf("Flannel (Ver: %s) is a simple and easy way to configure a layer 3 network fabric designed for Kubernetes.", version),
			PostInstall:     postInstall,
		},
	}, nil
}
