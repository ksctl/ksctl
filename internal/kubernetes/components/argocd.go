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

func setArgocdComponentOverridings(p metadata.ComponentOverrides) (
	version string,
	url []string,
	postInstall string,
) {
	url = nil
	postInstall = ""

	_version, _noUI, _namespaceInstall := getArgocdComponentOverridings(p)

	version = getVersionIfItsNotNilAndLatest(_version, "stable")

	generateManifestUrl := func(ver string, path string) string {
		return fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/%s", ver, path)
	}

	defaultVals := func() {
		url = []string{
			generateManifestUrl(version, "manifests/install.yaml"),
		}
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
			url = []string{
				generateManifestUrl(version, "manifests/core-install.yaml"),
			}
			postInstall = fmt.Sprintf(`
https://argo-cd.readthedocs.io/en/%s/operator-manual/core/
`, version)
		}
	} else if _namespaceInstall != nil {
		if *_namespaceInstall {
			url = []string{
				generateManifestUrl(version, "manifests/crds/application-crd.yaml"),
				generateManifestUrl(version, "manifests/crds/appproject-crd.yaml"),
				generateManifestUrl(version, "manifests/crds/applicationset-crd.yaml"),
				generateManifestUrl(version, "manifests/namespace-install.yaml"),
			}
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

func ArgoCDStandardComponent(params metadata.ComponentOverrides) metadata.StackComponent {
	version, url, postInstall := setArgocdComponentOverridings(params)

	return metadata.StackComponent{
		Kubectl: &metadata.KubectlHandler{
			Namespace:       "argocd",
			CreateNamespace: true,
			Urls:            url,
			Version:         version,
			Metadata:        fmt.Sprintf("Argo CD (Ver: %s) is a declarative, GitOps continuous delivery tool for Kubernetes.", version),
			PostInstall:     postInstall,
		},
		HandlerType: metadata.ComponentTypeKubectl,
	}
}
