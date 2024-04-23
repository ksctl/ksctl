package kubernetes

import (
	"fmt"
)

func argoRolloutsData(ver string) Application {
	// 1.6.0
	return Application{
		Name:       "argo-rollouts",
		Url:        fmt.Sprintf("https://github.com/argoproj/argo-rollouts/releases/download/%s/install.yaml", ver),
		Namespace:  "argo-rollouts",
		Maintainer: "Dipankar Das",
		Version:    ver,
		Metadata:   fmt.Sprintf("Argo Rollouts (Ver: %s) is a Kubernetes controller and set of CRDs which provide advanced deployment capabilities such as blue-green, canary, canary analysis, experimentation, and progressive delivery features to Kubernetes.", ver),
		PostInstall: `
Commands to execute to access Argo-Rollouts
	$ kubectl argo rollouts version
	$ kubectl argo rollouts dashboard
and open http://localhost:3100/rollouts
		`,
		InstallType:   InstallKubectl,
		KubectlConfig: KubectlOptions{createNamespace: true},
	}
}
