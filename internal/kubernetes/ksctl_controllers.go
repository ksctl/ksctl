package kubernetes

import "fmt"

func applicationStackData(params applicationParams) ApplicationStack {
	if params.version == "latest" {
		params.version = "main"
	}
	return ApplicationStack{
		StackNameID: KsctlApplicationOperatorID,
		Maintainer:  "github@dipankardas011",
		components: []StackComponent{
			{
				handlerType: ComponentTypeKubectl,
				kubectl: &KubectlHandler{
					url:             fmt.Sprintf("https://raw.githubusercontent.com/ksctl/ksctl/%s/ksctl-components/manifests/controllers/application/deploy.yml", params.version),
					createNamespace: false,
					metadata:        fmt.Sprintf("Ksctl Application controller (Ver: %s)", params.version),
					postInstall: `
					As the controller and the crd are installed just need to apply application to be installed
							`,
				},
			},
		},
	}
}
