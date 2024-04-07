package kubernetes

func argocdData() Application {
	return Application{
		Name:       "argocd",
		Url:        "https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml",
		Namespace:  "argocd",
		Maintainer: "Dipankar Das",
		Version:    "Latest stable version",
		Metadata:   "Argo CD is a declarative, GitOps continuous delivery tool for Kubernetes.",
		PostInstall: `
Commands to execute to access Argocd
	kubectl get secret -n argocd argocd-initial-admin-secret -o json | jq -r '.data.password' | base64 -d
	kubectl port-forward svc/argocd-server -n argocd 8080:443
and login to http://localhost:8080 with user admin and password from above
		`,
		InstallType:   InstallKubectl,
		KubectlConfig: KubectlOptions{createNamespace: true},
	}
}
