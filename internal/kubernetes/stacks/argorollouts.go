package stacks

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/kubernetes/components"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func argoRolloutsStandardCICD(params metadata.ApplicationParams) metadata.ApplicationStack {
	url := fmt.Sprintf("https://github.com/argoproj/argo-rollouts/releases/%s/download/install.yaml", params.Version)
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

	return metadata.ApplicationStack{
		Components: map[metadata.StackComponentID]metadata.StackComponent{
			metadata.ArgorolloutsComponentID: components.ArgoRolloutsStandardComponent(metadata.ComponentParams{
				Url:         url,
				Version:     params.Version,
				PostInstall: postInstall,
			}),
		},
		StackNameID: metadata.ArgoRolloutsStandardStackID,
		Maintainer:  "github@dipankardas011",
	}
}
