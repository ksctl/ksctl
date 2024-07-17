package components

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func KsctlApplicationComponent(params metadata.ComponentParams) metadata.StackComponent {
	return metadata.StackComponent{
		HandlerType: metadata.ComponentTypeKubectl,
		Kubectl: &metadata.KubectlHandler{
			Url:             fmt.Sprintf("https://raw.githubusercontent.com/ksctl/ksctl/%s/ksctl-components/manifests/controllers/application/deploy.yml", params.Version),
			CreateNamespace: false,
			Metadata:        fmt.Sprintf("Ksctl Application controller (Ver: %s)", params.Version),
			PostInstall:     params.PostInstall,
		},
	}
}
