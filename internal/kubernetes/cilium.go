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
