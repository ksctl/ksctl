package kubernetes

import (
	"fmt"
)

func argoRolloutsData(ver string) Application {
	return Application{
		Name:        "argo-rollouts",
		Url:         fmt.Sprintf("https://github.com/argoproj/argo-rollouts/releases/download/%s/install.yaml", ver),
		Maintainer:  "Dipankar Das",
		Version:     ver,
		InstallType: InstallKubectl,
		KubectlConfig: KubectlOptions{
			metadata: fmt.Sprintf("Argo Rollouts (Ver: %s) is a Kubernetes controller and set of CRDs which provide advanced deployment capabilities such as blue-green, canary, canary analysis, experimentation, and progressive delivery features to Kubernetes.", ver),
			postInstall: `
			Commands to execute to access Argo-Rollouts
			$ kubectl argo rollouts version
			$ kubectl argo rollouts dashboard
			and open http://localhost:3100/rollouts
			`,
			createNamespace: true,
			namespace:       "argo-rollouts",
		},
	}
}
