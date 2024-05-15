package kubernetes

import "fmt"

func applicationStackData(ver string) Application {
	if ver == "latest" {
		ver = "main"
	}
	return Application{
		Name:        "ksctl-appplication-stack",
		Url:         fmt.Sprintf("https://raw.githubusercontent.com/ksctl/ksctl/%s/ksctl-components/manifests/controllers/application/deploy.yml", ver),
		Maintainer:  "Dipankar Das",
		Version:     ver,
		InstallType: InstallKubectl,
		KubectlConfig: KubectlOptions{
			metadata: fmt.Sprintf("Ksctl Application controller (Ver: %s)", ver),
			postInstall: `
			As the controller and the crd are installed just need to apply application to be installed
			`,
			createNamespace: false,
		},
	}
}
