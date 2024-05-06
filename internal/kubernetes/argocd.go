package kubernetes

import "fmt"

func argocdData(ver string) Application {
	return Application{
		Name:       "argocd",
		Url:        fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/install.yaml", ver),
		Maintainer: "Dipankar Das",
		Version:    ver,
		Metadata:   fmt.Sprintf("Argo CD (Ver: %s) is a declarative, GitOps continuous delivery tool for Kubernetes.", ver),
		PostInstall: `
Commands to execute to access Argocd
	$ kubectl get secret -n argocd argocd-initial-admin-secret -o json | jq -r '.data.password' | base64 -d
	$ kubectl port-forward svc/argocd-server -n argocd 8080:443
and login to http://localhost:8080 with user admin and password from above
		`,
		InstallType: InstallKubectl,
		KubectlConfig: KubectlOptions{
			createNamespace: true,
			namespace:       "argocd",
		},
	}
}
