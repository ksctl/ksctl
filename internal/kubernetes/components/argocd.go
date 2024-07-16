package components

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func ArgoCDStandardComponent(params metadata.ComponentParams) metadata.StackComponent {
	return metadata.StackComponent{
		Kubectl: &metadata.KubectlHandler{
			Namespace:       "argocd",
			CreateNamespace: true,
			Url:             params.Url,
			Version:         params.Version,
			Metadata:        fmt.Sprintf("Argo CD (Ver: %s) is a declarative, GitOps continuous delivery tool for Kubernetes.", params.Version),
			PostInstall:     params.PostInstall,
		},
		HandlerType: metadata.ComponentTypeKubectl,
	}
}
