package kubernetes

import "fmt"

func ciliumData(ver string) Application {
	return Application{
		Name:       "cilium",
		Namespace:  "<Helm-Managed>",
		Url:        "https://helm.cilium.io/",
		Maintainer: "Dipankar Das",
		Version:    ver,
		Metadata:   fmt.Sprintf("Cilium (Ver: %s) is an open source, cloud native solution for providing, securing, and observing network connectivity between workloads, fueled by the revolutionary Kernel technology eBPF", ver),
		PostInstall: `
Once all the components feel ready
	$ cilium status
to check the status of the cilium installed
		`,
		InstallType: InstallHelm,
		HelmConfig: []HelmOptions{
			HelmOptions{
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
