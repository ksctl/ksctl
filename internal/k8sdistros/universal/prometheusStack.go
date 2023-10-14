package universal

func prometheusStackData() Application {
	return Application{
		Name:       "prometheus-community",
		Namespace:  "<Helm-Managed>",
		Url:        "https://prometheus-community.github.io/helm-charts",
		Maintainer: "Dipankar Das",
		Version:    "1.16.1",
		Metadata:   "Istio extends Kubernetes to establish a programmable, application-aware network using the powerful Envoy service proxy. Working with both Kubernetes and traditional workloads, Istio brings standard, universal traffic management, telemetry, and security to complex deployments.",
		PostInstall: `
TODO: Its blank
		`,
		InstallType: InstallHelm,
		HelmConfig: []WorkLoad{
			WorkLoad{
				chartName:   "prometheus-community/kube-prometheus-stack",
				chartVer:    "0.68.0",
				releaseName: "kube-prometheus-stack",
				// namespace:   "istio-system",
				// createns: true,
				// args: map[string]interface{}{
				// 	"defaultRevision": "default",
				// },
			},
		},
	}
}
