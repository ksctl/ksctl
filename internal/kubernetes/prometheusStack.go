package kubernetes

func prometheusStackData() Application {
	return Application{
		Name:       "prometheus-community",
		Namespace:  "<Helm-Managed>",
		Url:        "https://prometheus-community.github.io/helm-charts",
		Maintainer: "Dipankar Das",
		Version:    "51.7.0",
		Metadata:   "kube-prometheus-stack collects Kubernetes manifests, Grafana dashboards, and Prometheus rules combined with documentation and scripts to provide easy to operate end-to-end Kubernetes cluster monitoring with Prometheus using the Prometheus Operator.",
		PostInstall: `
credential for accessing grafana
	username: admin
	password: kube-operator
Done
		`,
		InstallType: InstallHelm,
		HelmConfig: []HelmOptions{
			HelmOptions{
				chartName:       "prometheus-community/kube-prometheus-stack",
				chartVer:        "51.7.0",
				releaseName:     "kube-prometheus-stack",
				namespace:       "monitoring",
				createNamespace: true,
			},
		},
	}
}
