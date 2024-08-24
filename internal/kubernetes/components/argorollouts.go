package components

import (
	"fmt"

	"github.com/ksctl/ksctl/pkg/helpers/utilities"
	"github.com/ksctl/ksctl/poller"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func getArgorolloutsComponentOverridings(p metadata.ComponentOverrides) (version *string, namespaceInstall *bool) {
	if p == nil {
		return nil, nil
	}

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "namespaceInstall":
			if v, ok := v.(bool); ok {
				namespaceInstall = utilities.Ptr(v)
			}
		}
	}
	return
}

func setArgorolloutsComponentOverridings(params metadata.ComponentOverrides) (
	version string,
	url string,
	postInstall string,
	err error,
) {
	releases, err := poller.GetSharedPoller().Get("argoproj", "argo-rollouts")
	if err != nil {
		return
	}

	url = ""
	postInstall = ""

	_version, _namespaceInstall := getArgorolloutsComponentOverridings(params)
	version = getVersionIfItsNotNilAndLatest(_version, releases[0])

	defaultVals := func() {
		url = fmt.Sprintf("https://github.com/argoproj/argo-rollouts/releases/download/%s/install.yaml", version)
		postInstall = `
Commands to execute to access Argo-Rollouts
$ kubectl argo rollouts version
$ kubectl argo rollouts dashboard
and open http://localhost:3100/rollouts
`
	}

	if _namespaceInstall != nil {
		if *_namespaceInstall {
			// TODO: need to install the crd as well
			url = fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-rollouts/%s/manifests/namespace-install.yaml", version)
			postInstall = fmt.Sprintf(`
https://argo-rollouts.readthedocs.io/en/%v/installation/#controller-installation
`, version)

		} else {
			defaultVals()
		}
	} else {
		defaultVals()
	}
	return
}

func ArgoRolloutsStandardComponent(params metadata.ComponentOverrides) (metadata.StackComponent, error) {
	version, url, postInstall, err := setArgorolloutsComponentOverridings(params)
	if err != nil {
		return metadata.StackComponent{}, err
	}

	return metadata.StackComponent{
		Kubectl: &metadata.KubectlHandler{
			Namespace:       "argo-rollouts",
			CreateNamespace: true,
			Url:             url,
			Version:         version,
			Metadata:        fmt.Sprintf("Argo Rollouts (Ver: %s) is a Kubernetes controller and set of CRDs which provide advanced deployment capabilities such as blue-green, canary, canary analysis, experimentation, and progressive delivery features to Kubernetes.", version),
			PostInstall:     postInstall,
		},
		HandlerType: metadata.ComponentTypeKubectl,
	}, nil
}
