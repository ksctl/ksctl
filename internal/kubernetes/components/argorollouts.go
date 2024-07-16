package components

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func ArgoRolloutsStandardComponent(params metadata.ComponentParams) metadata.StackComponent {
	return metadata.StackComponent{
		Kubectl: &metadata.KubectlHandler{
			Namespace:       "argo-rollouts",
			CreateNamespace: true,
			Url:             params.Url,
			Version:         params.Version,
			Metadata:        fmt.Sprintf("Argo Rollouts (Ver: %s) is a Kubernetes controller and set of CRDs which provide advanced deployment capabilities such as blue-green, canary, canary analysis, experimentation, and progressive delivery features to Kubernetes.", params.Version),
			PostInstall:     params.PostInstall,
		},
		HandlerType: metadata.ComponentTypeKubectl,
	}
}
