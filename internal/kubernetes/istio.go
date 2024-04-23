package kubernetes

import "fmt"

func istioData(ver string) Application {
	//1.16.1
	return Application{
		Name:       "istio",
		Namespace:  "<Helm-Managed>",
		Url:        "https://istio-release.storage.googleapis.com/charts",
		Maintainer: "Dipankar Das",
		Version:    ver,
		Metadata:   fmt.Sprintf("Istio (Ver: %s) extends Kubernetes to establish a programmable, application-aware network using the powerful Envoy service proxy. Working with both Kubernetes and traditional workloads, Istio brings standard, kubernetes traffic management, telemetry, and security to complex deployments.", ver),
		PostInstall: `
TODO: Its blank
		`,
		InstallType: InstallHelm,
		HelmConfig: []HelmOptions{
			HelmOptions{
				chartName:       "istio/base",
				chartVer:        ver,
				releaseName:     "istio-base",
				namespace:       "istio-system",
				createNamespace: true,
				args: map[string]interface{}{
					"defaultRevision": "default",
				},
			},
			HelmOptions{
				chartName:       "istio/istiod",
				chartVer:        ver,
				releaseName:     "istiod",
				namespace:       "istio-system",
				createNamespace: false,
				args:            nil,
			},
		},
	}
}
