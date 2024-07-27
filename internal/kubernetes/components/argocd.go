package components

import (
	"fmt"

	"github.com/ksctl/ksctl/pkg/helpers/utilities"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func getArgocdComponentOverridings(p metadata.ComponentOverrides) (version *string, noUI *bool, namespaceInstall *bool) {
	if p == nil {
		return nil, nil, nil
	}
	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "noUI":
			if v, ok := v.(bool); ok {
				noUI = utilities.Ptr(v)
			}
		case "namespaceInstall":
			if v, ok := v.(bool); ok {
				namespaceInstall = utilities.Ptr(v)
			}
		}
	}
	return
}

func ArgoCDStandardComponent(params metadata.ComponentOverrides) metadata.StackComponent {
	var (
		version     = "stable"
		url         = ""
		postInstall = ""
	)
	_version, _noUI, _namespaceInstall := getArgocdComponentOverridings(params)
	if _version != nil {
		if *_version != "latest" {
			version = *_version
		}
	}

	defaultVals := func() {
		url = fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/install.yaml", version)
		postInstall = `
Commands to execute to access Argocd
$ kubectl get secret -n argocd argocd-initial-admin-secret -o json | jq -r '.data.password' | base64 -d
$ kubectl port-forward svc/argocd-server -n argocd 8080:443
and login to http://localhost:8080 with user admin and password from above
`
	}

	if _noUI != nil {
		if *_noUI {
			defaultVals()
		} else {
			url = fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/core-install.yaml", version)
			postInstall = fmt.Sprintf(`
https://argo-cd.readthedocs.io/en/%s/operator-manual/core/
`, version)
		}
	} else if _namespaceInstall != nil {
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

	return metadata.StackComponent{
		Kubectl: &metadata.KubectlHandler{
			Namespace:       "argocd",
			CreateNamespace: true,
			Url:             url,
			Version:         version,
			Metadata:        fmt.Sprintf("Argo CD (Ver: %s) is a declarative, GitOps continuous delivery tool for Kubernetes.", version),
			PostInstall:     postInstall,
		},
		HandlerType: metadata.ComponentTypeKubectl,
	}
}
