package universal

func ciliumData() Application {
	return Application{
		Name:       "cilium",
		Namespace:  "<Helm-Managed>",
		Url:        "https://helm.cilium.io/",
		Maintainer: "Dipankar Das",
		Version:    "1.14.2",
		Metadata:   "Cilium is an open source, cloud native solution for providing, securing, and observing network connectivity between workloads, fueled by the revolutionary Kernel technology eBPF",
		PostInstall: `
Once all the components feel ready
	cilium status
to check the status of the cilium installed
		`,
		InstallType: InstallHelm,
		HelmConfig: []WorkLoad{
			WorkLoad{
				chartName:   "cilium/cilium",
				chartVer:    "1.14.2",
				releaseName: "cilium",
				namespace:   "kube-system",
				createns:    false,
				args:        nil,
			},
		},
	}
}
