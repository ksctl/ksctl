package kubernetes

func argoRolloutsData() Application {
	return Application{
		Name:       "argo-rollouts",
		Url:        "https://github.com/argoproj/argo-rollouts/releases/download/v1.6.0/install.yaml",
		Namespace:  "argo-rollouts",
		Maintainer: "Dipankar Das",
		Version:    "v1.6.0",
		Metadata:   "Argo Rollouts is a Kubernetes controller and set of CRDs which provide advanced deployment capabilities such as blue-green, canary, canary analysis, experimentation, and progressive delivery features to Kubernetes.",
		PostInstall: `
Commands to execute to access Argo-Rollouts
	kubectl argo rollouts version
	kubectl argo rollouts dashboard
and open http://localhost:3100/rollouts
		`,
		InstallType:   InstallKubectl,
		KubectlConfig: KubectlOptions{createNamespace: true},
	}
}
