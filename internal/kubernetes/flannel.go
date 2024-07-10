package kubernetes

import "fmt"

func flannelStandardCNI(params applicationParams) ApplicationStack {
	if params.version == "stable" {
		params.version = "latest"
	}
	return ApplicationStack{
		Maintainer:  "github:dipankardas011",
		StackNameID: FlannelStandardStackID,
		components: []StackComponent{
			{
				handlerType: ComponentTypeKubectl,
				kubectl: &KubectlHandler{
					url:             fmt.Sprintf("https://github.com/flannel-io/flannel/releases/%s/download/kube-flannel.yml", params.version),
					version:         params.version,
					createNamespace: false,
					metadata:        fmt.Sprintf("Flannel (Ver: %s) is a simple and easy way to configure a layer 3 network fabric designed for Kubernetes.", params.version),
					postInstall: `
	None
			`,
				},
			},
		},
	}
}
