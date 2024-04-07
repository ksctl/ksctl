package kubernetes

func istioData() Application {
	return Application{
		Name:       "istio",
		Namespace:  "<Helm-Managed>",
		Url:        "https://istio-release.storage.googleapis.com/charts",
		Maintainer: "Dipankar Das",
		Version:    "1.16.1",
		Metadata:   "Istio extends Kubernetes to establish a programmable, application-aware network using the powerful Envoy service proxy. Working with both Kubernetes and traditional workloads, Istio brings standard, kubernetes traffic management, telemetry, and security to complex deployments.",
		PostInstall: `
TODO: Its blank
		`,
		InstallType: InstallHelm,
		HelmConfig: []HelmOptions{
			HelmOptions{
				chartName:       "istio/base",
				chartVer:        "1.16.1",
				releaseName:     "istio-base",
				namespace:       "istio-system",
				createNamespace: true,
				args: map[string]interface{}{
					"defaultRevision": "default",
				},
			},
			HelmOptions{
				chartName:       "istio/istiod",
				chartVer:        "1.16.1",
				releaseName:     "istiod",
				namespace:       "istio-system",
				createNamespace: false,
				args:            nil,
			},
		},
	}
}
