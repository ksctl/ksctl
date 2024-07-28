package components

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
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
) {
	version = "latest"
	url = ""
	postInstall = ""

	_version := getFlannelComponentOverridings(p)
	if _version != nil {
		version = *_version
	}

	defaultVals := func() {
		url = fmt.Sprintf("https://github.com/flannel-io/flannel/releases/%s/download/kube-flannel.yml", version)
		postInstall = "https://github.com/flannel-io/flannel"
	}

	defaultVals()
	return
}

func FlannelStandardComponent(params metadata.ComponentOverrides) metadata.StackComponent {

	version, url, postInstall := setFlannelComponentOverridings(params)
	return metadata.StackComponent{
		HandlerType: metadata.ComponentTypeKubectl,
		Kubectl: &metadata.KubectlHandler{
			Url:             url,
			Version:         version,
			CreateNamespace: false,
			Metadata:        fmt.Sprintf("Flannel (Ver: %s) is a simple and easy way to configure a layer 3 network fabric designed for Kubernetes.", version),
			PostInstall:     postInstall,
		},
	}
}
