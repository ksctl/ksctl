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

func argoRolloutsStandardApp(ver string, clusterAccess bool) ApplicationStack {
	url := fmt.Sprintf("https://github.com/argoproj/argo-rollouts/releases/%s/download/install.yaml", ver)
	postInstall := `
	Commands to execute to access Argo-Rollouts
	$ kubectl argo rollouts version
	$ kubectl argo rollouts dashboard
	and open http://localhost:3100/rollouts
`
	// TODO: need to make the namespace based install work
	// Refer: https://argoproj.github.io/argo-rollouts/installation/#controller-installation

	// 	if !clusterAccess {
	// 		url = fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/namespace-install.yaml", ver)
	// 		postInstall = `
	// https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/#non-high-availability
	// `
	// 	}

	return ApplicationStack{
		components: []StackComponent{
			{
				kubectl: &KubectlHandler{
					namespace:       "argo-rollouts",
					createNamespace: true,
					url:             url,
					version:         ver,
					metadata:        fmt.Sprintf("Argo Rollouts (Ver: %s) is a Kubernetes controller and set of CRDs which provide advanced deployment capabilities such as blue-green, canary, canary analysis, experimentation, and progressive delivery features to Kubernetes.", ver),
					postInstall:     postInstall,
				},
				handlerType: ComponentTypeKubectl,
			},
		},
		StackType:   StackTypeStandard,
		StackNameID: ArgocdStandardStackID,
		Maintainer:  "github@dipankardas011",
	}
}
