package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func KubespinProductionApp(params metadata.ApplicationParams) metadata.ApplicationStack {
	return metadata.ApplicationStack{
		Maintainer:  "github@dipankardas011",
		StackNameID: metadata.KubeSpinProductionStackID,
		Components: map[metadata.StackComponentID]metadata.StackComponent{
			metadata.CertManagerComponentID: components.CertManagerComponent(
				params.ComponentParams[metadata.CertManagerComponentID],
			),
			metadata.KwasmOperatorComponentID: components.KwasmOperatorComponent(
				params.ComponentParams[metadata.KwasmOperatorComponentID],
			),
		},
		StkDepsIdx: []metadata.StackComponentID{
			metadata.CertManagerComponentID,
			metadata.KwasmOperatorComponentID,
		},
		// components: []kubernetes.StackComponent{
		// 	{
		// 		handlerType: kubernetes.ComponentTypeKubectl,
		// 		kubectl: &kubernetes.KubectlHandler{
		// 			createNamespace: false,
		// 			url:             fmt.Sprintf("https://github.com/spinkube/spin-operator/releases/download/v0.2.0/spin-operator.crds.yaml"),
		// 			version:         ver,
		// 			metadata:        `TODO`,
		// 			postInstall:     `TODO`,
		// 		},
		// 	},
		// 	{
		// 		handlerType: kubernetes.ComponentTypeKubectl,
		// 		kubectl: &kubernetes.KubectlHandler{
		// 			createNamespace: false,
		// 			url:             fmt.Sprintf("https://github.com/spinkube/spin-operator/releases/download/v0.2.0/spin-operator.runtime-class.yaml"),
		// 			version:         ver,
		// 			metadata:        `TODO`,
		// 			postInstall:     `TODO`,
		// 		},
		// 	},
		// 	{
		// 		handlerType: kubernetes.ComponentTypeKubectl,
		// 		kubectl: &kubernetes.KubectlHandler{
		// 			createNamespace: false,
		// 			url:             fmt.Sprintf("https://github.com/spinkube/spin-operator/releases/download/v0.2.0/spin-operator.shim-executor.yaml"),
		// 			version:         ver,
		// 			metadata:        `TODO`,
		// 			postInstall:     `TODO`,
		// 		},
		// 	},
		// 	{
		// 		handlerType: kubernetes.ComponentTypeHelm,
		// 		helm:        &kubernetes.HelmHandler{
		// 			// Not sure how it is interepreseted
		// 			//helm install spin-operator \
		// 			//--namespace spin-operator \
		// 			//--create-namespace \
		// 			//--version 0.2.0 \
		// 			//--wait \
		// 			//oci://ghcr.io/spinkube/charts/spin-operator
		// 		},
		// 	},
		// 	// # Uninstall Spin Operator using Helm
		// 	//helm delete spin-operator --namespace spin-operator
		// 	//kubectl delete -f https://github.com/spinkube/spin-operator/releases/download/v0.2.0/spin-operator.shim-executor.yaml
		// 	//kubectl delete -f https://github.com/spinkube/spin-operator/releases/download/v0.2.0/spin-operator.runtime-class.yaml
		// 	//kubectl delete -f https://github.com/spinkube/spin-operator/releases/download/v0.2.0/spin-operator.crds.yaml
		// },
	}
}
