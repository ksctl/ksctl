package kubernetes

import "fmt"

func flannelData(ver string) Application {
	if ver == "stable" {
		ver = "latest"
	}
	return Application{
		Name:       "flannel",
		Url:        fmt.Sprintf("https://github.com/flannel-io/flannel/releases/%s/download/kube-flannel.yml", ver),
		Namespace:  "flannel",
		Maintainer: "Dipankar Das",
		Version:    ver,
		Metadata:   fmt.Sprintf("Flannel (Ver: %s) is a simple and easy way to configure a layer 3 network fabric designed for Kubernetes.", ver),
		PostInstall: `
None
		`,
		InstallType:   InstallKubectl,
		KubectlConfig: KubectlOptions{createNamespace: false},
	}
}
