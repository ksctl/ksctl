package components

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

func getSpinkubeComponentOverridings(p metadata.ComponentOverrides) (version *string) {
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

func setSpinkubeComponentOverridings(p metadata.ComponentOverrides) (
	version string,
	url string,
	postInstall string,
	err error,
) {

	version, err = utilities.GetLatestRepoRelease("spinkube", "spin-operator")
	if err != nil {
		return
	}

	url = ""
	postInstall = ""

	_version := getSpinkubeComponentOverridings(p)
	if _version != nil {
		version = *_version
	}

	defaultVals := func() {
		url = fmt.Sprintf("https://github.com/spinkube/spin-operator/releases/download/%s/spin-operator.crds.yaml", version)
		postInstall = "https://www.spinkube.dev/docs/topics/"
	}

	defaultVals()
	return
}

func SpinkubeStandardComponent(params metadata.ComponentOverrides) (metadata.StackComponent, error) {

	version, url, postInstall, err := setSpinkubeComponentOverridings(params)
	if err != nil {
		return metadata.StackComponent{}, err
	}

	return metadata.StackComponent{
		HandlerType: metadata.ComponentTypeKubectl,
		Kubectl: &metadata.KubectlHandler{
			Url:             url,
			Version:         version,
			CreateNamespace: false,
			Metadata:        fmt.Sprintf("KubeSpin (ver: %s) is an open source project that streamlines developing, deploying and operating WebAssembly workloads in Kubernetes - resulting in delivering smaller, more portable applications and incredible compute performance benefits", version),
			PostInstall:     postInstall,
		},
	}, nil
}
