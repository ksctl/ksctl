package kubernetes

func storageImportData() Application {
	return Application{
		Name:       "ksctl-import-data",
		Url:        "https://raw.githubusercontent.com/ksctl/ksctl/251-sub-feature-ksctl-agent/ksctl-components/manifests/controllers/storage/deploy.yml",
		Namespace:  "ksctl",
		Maintainer: "Dipankar Das",
		Version:    "Alpha version",
		Metadata:   "It Installs the controller",
		PostInstall: `
As the controller and the crd are installed just need to apply storage data exported before
		`,
		InstallType:   InstallKubectl,
		KubectlConfig: KubectlOptions{createNamespace: false},
	}
}
