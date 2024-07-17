package components

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func FlannelStandardComponent(params metadata.ComponentParams) metadata.StackComponent {
	return metadata.StackComponent{
		HandlerType: metadata.ComponentTypeKubectl,
		Kubectl: &metadata.KubectlHandler{
			Url:             fmt.Sprintf("https://github.com/flannel-io/flannel/releases/%s/download/kube-flannel.yml", params.Version),
			Version:         params.Version,
			CreateNamespace: false,
			Metadata:        fmt.Sprintf("Flannel (Ver: %s) is a simple and easy way to configure a layer 3 network fabric designed for Kubernetes.", params.Version),
			PostInstall:     params.PostInstall,
		},
	}
}
