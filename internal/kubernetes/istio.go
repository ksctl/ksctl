package kubernetes

func istioData(ver string) Application {
	return Application{
		Name:        "istio",
		Url:         "https://istio-release.storage.googleapis.com/charts",
		Maintainer:  "Dipankar Das",
		Version:     ver,
		InstallType: InstallHelm,
		HelmConfig: []HelmOptions{
			{
				chartName:       "istio/base",
				chartVer:        ver,
				releaseName:     "istio-base",
				namespace:       "istio-system",
				createNamespace: true,
				args: map[string]interface{}{
					"defaultRevision": "default",
				},
			},
			{
				chartName:       "istio/istiod",
				chartVer:        ver,
				releaseName:     "istiod",
				namespace:       "istio-system",
				createNamespace: false,
				args:            nil,
			},
		},
	}
}

func istioStandardServiceMesh(params applicationParams) ApplicationStack {
	return ApplicationStack{
		components: []StackComponent{
			{
				helm: &HelmHandler{
					repoUrl:  "https://istio-release.storage.googleapis.com/charts",
					repoName: "istio",
					charts: []HelmOptions{
						{
							chartName:       "istio/base",
							chartVer:        params.version,
							releaseName:     "istio-base",
							namespace:       "istio-system",
							createNamespace: true,
							args: map[string]interface{}{
								"defaultRevision": "default",
							},
						},
						{
							chartName:       "istio/istiod",
							chartVer:        params.version,
							releaseName:     "istiod",
							namespace:       "istio-system",
							createNamespace: false,
							args:            nil,
						},
					},
				},
				handlerType: ComponentTypeHelm,
			},
		},
		Maintainer:  "github:dipankardas011",
		StackNameID: IstioStandardStackID,
	}
}
