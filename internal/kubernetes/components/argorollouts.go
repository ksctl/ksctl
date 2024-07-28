package components

import (
	"fmt"

	"github.com/ksctl/ksctl/pkg/helpers/utilities"

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
) {
	version = "latest"
	url = ""
	postInstall = ""

	_version, _namespaceInstall := getArgorolloutsComponentOverridings(params)
	if _version != nil {
		version = *_version
	}

	defaultVals := func() {
		url = fmt.Sprintf("https://github.com/argoproj/argo-rollouts/releases/%s/download/install.yaml", version)
		postInstall = `
Commands to execute to access Argo-Rollouts
$ kubectl argo rollouts version
$ kubectl argo rollouts dashboard
and open http://localhost:3100/rollouts
`
	}

	if _namespaceInstall != nil {
		if *_namespaceInstall {
			url = fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/namespace-install.yaml", version)
			postInstall = fmt.Sprintf(`
https://argo-cd.readthedocs.io/en/%s/operator-manual/installation/#non-high-availability
`, version)

		} else {
			defaultVals()
		}
	} else {
		defaultVals()
	}
	return
}

func ArgoRolloutsStandardComponent(params metadata.ComponentOverrides) metadata.StackComponent {
	version, url, postInstall := setArgorolloutsComponentOverridings(params)

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
	}
}
