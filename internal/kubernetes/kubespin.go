package kubernetes

import "fmt"

// TODO(dipankar): can utilize multiple apps rather than having them defined multiple places
// as it makes more sense
func kubespinProductionApp(ver string) ApplicationStack {
	return ApplicationStack{
		Maintainer:  "github@dipankardas011",
		StackType:   StackTypeProduction,
		StackNameID: KubeSpinProductionStackID,
		components: []StackComponent{
			{
				kubectl: &KubectlHandler{
					createNamespace: false,
					url:             fmt.Sprintf("https://github.com/cert-manager/cert-manager/releases/download/v1.14.5/cert-manager.crds.yaml"),
					version:         ver,
					metadata:        `TODO`,
					postInstall:     `TODO`,
				},
				handlerType: ComponentTypeKubectl,
			},
			{
				kubectl: &KubectlHandler{
					createNamespace: false,
					url:             fmt.Sprintf("https://github.com/cert-manager/cert-manager/releases/download/v1.14.5/cert-manager.yaml"),
					version:         ver,
					metadata:        `TODO`,
					postInstall:     `TODO`,
				},
				handlerType: ComponentTypeKubectl,
			},
			{
				helm: &HelmHandler{
					repoUrl:  "http://kwasm.sh/kwasm-operator/",
					repoName: "kwasm",
					charts: []HelmOptions{
						{
							chartName:       "kwasm/kwasm-operator",
							chartVer:        ver,
							releaseName:     "kwasm-operator",
							namespace:       "kwasm",
							createNamespace: true,
							args: map[string]interface{}{
								"kwasmOperator.installerImage": "ghcr.io/spinkube/containerd-shim-spin/node-installer:v0.15.0",
							},
						},
					},
				},
				handlerType: ComponentTypeHelm,
			},
			{
				handlerType: ComponentTypeKubectl,
				kubectl: &KubectlHandler{
					createNamespace: false,
					url:             fmt.Sprintf("https://github.com/spinkube/spin-operator/releases/download/v0.2.0/spin-operator.crds.yaml"),
					version:         ver,
					metadata:        `TODO`,
					postInstall:     `TODO`,
				},
			},
			{
				handlerType: ComponentTypeKubectl,
				kubectl: &KubectlHandler{
					createNamespace: false,
					url:             fmt.Sprintf("https://github.com/spinkube/spin-operator/releases/download/v0.2.0/spin-operator.runtime-class.yaml"),
					version:         ver,
					metadata:        `TODO`,
					postInstall:     `TODO`,
				},
			},
			{
				handlerType: ComponentTypeKubectl,
				kubectl: &KubectlHandler{
					createNamespace: false,
					url:             fmt.Sprintf("https://github.com/spinkube/spin-operator/releases/download/v0.2.0/spin-operator.shim-executor.yaml"),
					version:         ver,
					metadata:        `TODO`,
					postInstall:     `TODO`,
				},
			},
			{
				handlerType: ComponentTypeHelm,
				helm:        &HelmHandler{
					// Not sure how it is interepreseted
					//helm install spin-operator \
					//--namespace spin-operator \
					//--create-namespace \
					//--version 0.2.0 \
					//--wait \
					//oci://ghcr.io/spinkube/charts/spin-operator
				},
			},
			// # Uninstall Spin Operator using Helm
			//helm delete spin-operator --namespace spin-operator
			//kubectl delete -f https://github.com/spinkube/spin-operator/releases/download/v0.2.0/spin-operator.shim-executor.yaml
			//kubectl delete -f https://github.com/spinkube/spin-operator/releases/download/v0.2.0/spin-operator.runtime-class.yaml
			//kubectl delete -f https://github.com/spinkube/spin-operator/releases/download/v0.2.0/spin-operator.crds.yaml
		},
	}
}
