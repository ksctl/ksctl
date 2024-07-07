package kubernetes

import "fmt"

func flannelData(ver string) Application {
	if ver == "stable" {
		ver = "latest"
	}
	return Application{
		Name:        "flannel",
		Url:         fmt.Sprintf("https://github.com/flannel-io/flannel/releases/%s/download/kube-flannel.yml", ver),
		Maintainer:  "Dipankar Das",
		Version:     ver,
		InstallType: InstallKubectl,
		KubectlConfig: KubectlOptions{
			createNamespace: false,
			metadata:        fmt.Sprintf("Flannel (Ver: %s) is a simple and easy way to configure a layer 3 network fabric designed for Kubernetes.", ver),
			postInstall: `
	None
			`,
		},
	}
}

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
