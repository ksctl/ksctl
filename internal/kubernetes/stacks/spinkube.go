package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func SpinkubeProductionApp(params metadata.ApplicationParams) (metadata.ApplicationStack, error) {

	certManagerComponent, err := components.CertManagerComponent(
		params.ComponentParams[metadata.CertManagerComponentID],
	)
	if err != nil {
		return metadata.ApplicationStack{}, err
	}

	kwasmOperatorComponent, err := components.KwasmOperatorComponent(
		params.ComponentParams[metadata.KwasmOperatorComponentID],
	)
	if err != nil {
		return metadata.ApplicationStack{}, err
	}

	spinKubeCrd, err := components.SpinkubeOperatorCrdComponent(
		params.ComponentParams[metadata.SpinkubeOperatorCrdComponentID],
	)
	if err != nil {
		return metadata.ApplicationStack{}, err
	}

	spinKubeRuntime, err := components.SpinkubeOperatorRuntimeClassComponent(
		params.ComponentParams[metadata.SpinKubeOperatorRuntimeClassID],
	)
	if err != nil {
		return metadata.ApplicationStack{}, err
	}

	spinKubeShim, err := components.SpinkubeOperatorShimExecComponent(
		params.ComponentParams[metadata.SpinKubeOperatorShimExecutorID],
	)
	if err != nil {
		return metadata.ApplicationStack{}, err
	}

	spinOperator, err := components.SpinOperatorComponent(
		params.ComponentParams[metadata.SpinKubeOperatorComponentID],
	)
	if err != nil {
		return metadata.ApplicationStack{}, err
	}

	// TODO: need to make delete work perfrrectly
	//  Ctx: stkDepsIdx mode of traversal
	//  Install:
	//   - idx left to right
	//  Uninstall:
	//   - idx right to left

	return metadata.ApplicationStack{
		Maintainer:  "github@dipankardas011",
		StackNameID: metadata.SpinKubeProductionStackID,
		Components: map[metadata.StackComponentID]metadata.StackComponent{
			metadata.CertManagerComponentID:         certManagerComponent,
			metadata.KwasmOperatorComponentID:       kwasmOperatorComponent,
			metadata.SpinkubeOperatorCrdComponentID: spinKubeCrd,
			metadata.SpinKubeOperatorRuntimeClassID: spinKubeRuntime,
			metadata.SpinKubeOperatorShimExecutorID: spinKubeShim,
			metadata.SpinKubeOperatorComponentID:    spinOperator,
		},
		StkDepsIdx: []metadata.StackComponentID{
			metadata.CertManagerComponentID,
			metadata.KwasmOperatorComponentID,
			metadata.SpinkubeOperatorCrdComponentID,
			metadata.SpinKubeOperatorRuntimeClassID,
			metadata.SpinKubeOperatorShimExecutorID,
			metadata.SpinKubeOperatorComponentID,
		},
	}, nil
}

// components: []kubernetes.StackComponent{
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
