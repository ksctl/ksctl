package kubernetes

func ciliumStandardCNI(params applicationParams) ApplicationStack {
	return ApplicationStack{
		components: []StackComponent{
			{
				helm: &HelmHandler{
					repoName: "cilium",
					repoUrl:  "https://helm.cilium.io/",
					charts: []HelmOptions{
						{
							chartName:       "cilium/cilium",
							chartVer:        params.version,
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
