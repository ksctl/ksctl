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
			version = utilities.Ptr(v.(string))
		}
	}
	return
}

func KsctlApplicationComponent(params metadata.ComponentOverrides) metadata.StackComponent {
	var (
		version     = "main" // latest -> main
		postInstall = ""
		url         = ""
	)

	_version := getKsctlApplicationComponentOverridings(params)
	if _version != nil {
		if *_version != "latest" {
			version = *_version
		}
	}

	postInstall = "As the controller and the crd are installed just need to apply application to be installed"
	url = fmt.Sprintf("https://raw.githubusercontent.com/ksctl/ksctl/%s/ksctl-components/manifests/controllers/application/deploy.yml", version)

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
