package components

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

func getKsctlApplicationComponentOverridings(p metadata.ComponentOverrides) (version *string) {
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

func setKsctlApplicationComponentOverridings(p metadata.ComponentOverrides) (
	version string,
	url string,
	postInstall string,
) {
	version = "main"

	_version := getKsctlApplicationComponentOverridings(p)
	if _version != nil {
		if *_version != "latest" {
			version = *_version
		}
	}

	postInstall = "As the controller and the crd are installed just need to apply application to be installed"
	url = fmt.Sprintf("https://raw.githubusercontent.com/ksctl/ksctl/%s/ksctl-components/manifests/controllers/application/deploy.yml", version)

	return
}

func KsctlApplicationComponent(params metadata.ComponentOverrides) metadata.StackComponent {
	version, url, postInstall := setKsctlApplicationComponentOverridings(params)

	return metadata.StackComponent{
		HandlerType: metadata.ComponentTypeKubectl,
		Kubectl: &metadata.KubectlHandler{
			Url:             url,
			CreateNamespace: false,
			Metadata:        fmt.Sprintf("Ksctl Application controller (Ver: %s)", version),
			PostInstall:     postInstall,
		},
	}
}
