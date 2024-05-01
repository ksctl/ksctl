package kubernetes

import "fmt"

func istioData(ver string) Application {
	return Application{
		Name:       "istio",
		Namespace:  "<Helm-Managed>",
		Url:        "https://istio-release.storage.googleapis.com/charts",
		Maintainer: "Dipankar Das",
		Version:    ver,
		Metadata:   fmt.Sprintf("Istio (Ver: %s) extends Kubernetes to establish a programmable, application-aware network using the powerful Envoy service proxy. Working with both Kubernetes and traditional workloads, Istio brings standard, kubernetes traffic management, telemetry, and security to complex deployments.", ver),
		PostInstall: `
"istiod" successfully installed!

To learn more about the release, try:
  $ helm status istiod
  $ helm get all istiod

Next steps:
  * Deploy a Gateway: https://istio.io/latest/docs/setup/additional-setup/gateway/
  * Try out our tasks to get started on common configurations:
    * https://istio.io/latest/docs/tasks/traffic-management
    * https://istio.io/latest/docs/tasks/security/
    * https://istio.io/latest/docs/tasks/policy-enforcement/
    * https://istio.io/latest/docs/tasks/policy-enforcement/
  * Review the list of actively supported releases, CVE publications and our hardening guide:
    * https://istio.io/latest/docs/releases/supported-releases/
    * https://istio.io/latest/news/security/
    * https://istio.io/latest/docs/ops/best-practices/security/

For further documentation see https://istio.io website

Tell us how your install/upgrade experience went at https://forms.gle/99uiMML96AmsXY5d6
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
