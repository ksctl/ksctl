package components

import (
	"fmt"

	"github.com/ksctl/ksctl/pkg/helpers/utilities"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func getArgorolloutsComponentOverridings(p metadata.ComponentOverriding) (version *string, namespaceInstall *bool) {
	if p == nil {
		return nil, nil
	}

	for k, v := range p {
		switch k {
		case "version":
			version = utilities.Ptr(v.(string))
		case "namespaceInstall":
			namespaceInstall = utilities.Ptr(v.(bool))
		}
	}
	return
}

func ArgoRolloutsStandardComponent(params metadata.ComponentOverriding) metadata.StackComponent {

	var (
		version     = "latest"
		url         = ""
		postInstall = ""
	)
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
			postInstall = `
https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/#non-high-availability
`

		} else {
			defaultVals()
		}
	} else {
		defaultVals()
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
	}
}
