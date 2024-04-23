package kubernetes

import "fmt"

func prometheusStackData(ver string) Application {
	//51.7.0

	return Application{
		Name:       "prometheus-community",
		Namespace:  "<Helm-Managed>",
		Url:        "https://prometheus-community.github.io/helm-charts",
		Maintainer: "Dipankar Das",
		Version:    ver,
		Metadata:   fmt.Sprintf("kube-prometheus-stack (Ver: %s) collects Kubernetes manifests, Grafana dashboards, and Prometheus rules combined with documentation and scripts to provide easy to operate end-to-end Kubernetes cluster monitoring with Prometheus using the Prometheus Operator.", ver),
		PostInstall: `
credential for accessing grafana
and open http://localhost:8080 by exposing the grafana service
	username: admin
	password: kube-operator
Done
		`,
		InstallType: InstallHelm,
		HelmConfig: []HelmOptions{
			HelmOptions{
				chartName:       "prometheus-community/kube-prometheus-stack",
				chartVer:        ver,
				releaseName:     "kube-prometheus-stack",
				namespace:       "monitoring",
				createNamespace: true,
			},
		},
	}
}
