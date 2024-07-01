package kubernetes

func ciliumData(ver string) Application {
	return Application{
		Name:        "cilium",
		Url:         "https://helm.cilium.io/",
		Maintainer:  "Dipankar Das",
		Version:     ver,
		InstallType: InstallHelm,
		HelmConfig: []HelmOptions{
			{
				chartName:       "cilium/cilium",
				chartVer:        ver,
				releaseName:     "cilium",
				namespace:       "kube-system",
				createNamespace: false,
				args:            nil,
			},
		},
	}
}

func ciliumStandardCNI(ver string) ApplicationStack {
	return ApplicationStack{
		components: []StackComponent{
			{
				helm: &HelmHandler{
					repoName: "cilium",
					repoUrl:  "https://helm.cilium.io/",
					charts: []HelmOptions{
						{
							chartName:       "cilium/cilium",
							chartVer:        ver,
							releaseName:     "cilium",
							namespace:       "kube-system",
							createNamespace: false,
							args:            nil,
						},
					},
				},
				handlerType: ComponentTypeHelm,
			},
		},
		StackNameID: CiliumStandardStackID,
		Maintainer:  "github@dipankardas011",
	}
}
