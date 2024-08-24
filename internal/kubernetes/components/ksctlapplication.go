package components

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

func getKsctlApplicationComponentOverridings(p metadata.ComponentOverrides) (version *string, uri *string) {
	if p == nil {
		return nil, nil
	}

	if v, ok := p["version"]; ok {
		if v, ok := v.(string); ok {
			version = utilities.Ptr(v)
		}
	}

	if v, ok := p["uri"]; ok {
		if v, ok := v.(string); ok {
			uri = utilities.Ptr(v)
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
	url = fmt.Sprintf(
		"https://raw.githubusercontent.com/ksctl/ksctl/%s/ksctl-components/manifests/controllers/application/deploy.yml",
		version)

	_version, _uri := getKsctlApplicationComponentOverridings(p)
	if _version != nil {
		version = *_version
	}

	if _uri != nil {
		url = *_uri
	}

	postInstall = "As the controller and the crd are installed just need to apply application to be installed"

	return
}

func KsctlApplicationComponent(params metadata.ComponentOverrides) metadata.StackComponent {
	version, url, postInstall := setKsctlApplicationComponentOverridings(params)

	return metadata.StackComponent{
		HandlerType: metadata.ComponentTypeKubectl,
		Kubectl: &metadata.KubectlHandler{
			Urls:            []string{url},
			CreateNamespace: false,
			Metadata:        fmt.Sprintf("Ksctl Application controller (Ver: %s)", version),
			PostInstall:     postInstall,
		},
	}
}
