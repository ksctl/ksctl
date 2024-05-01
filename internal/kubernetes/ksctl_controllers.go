package kubernetes

import "fmt"

func storageImportData(ver string) Application {
	if ver == "latest" {
		ver = "main"
	}
	return Application{
		Name:       "ksctl-import-data",
		Url:        fmt.Sprintf("https://raw.githubusercontent.com/ksctl/ksctl/%s/ksctl-components/manifests/controllers/storage/deploy.yml", ver),
		Namespace:  "ksctl",
		Maintainer: "Dipankar Das",
		Version:    ver,
		Metadata:   fmt.Sprintf("Ksctl Storage controller (Ver: %s)", ver),
		PostInstall: `
As the controller and the crd are installed just need to apply storage data exported before
		`,
		InstallType:   InstallKubectl,
		KubectlConfig: KubectlOptions{createNamespace: false},
	}
}

func applicationStackData(ver string) Application {
	if ver == "latest" {
		ver = "main"
	}
	return Application{
		Name:       "ksctl-appplication-stack",
		Url:        fmt.Sprintf("https://raw.githubusercontent.com/ksctl/ksctl/%s/ksctl-components/manifests/controllers/application/deploy.yml", ver),
		Namespace:  "ksctl",
		Maintainer: "Dipankar Das",
		Version:    ver,
		Metadata:   fmt.Sprintf("Ksctl Application controller (Ver: %s)", ver),
		PostInstall: `
As the controller and the crd are installed just need to apply application to be installed
		`,
		InstallType:   InstallKubectl,
		KubectlConfig: KubectlOptions{createNamespace: false},
	}
}
