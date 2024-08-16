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

	if err := components.GetSpinKubeStackSpecificKwasmOverrides(
		params.ComponentParams[metadata.KwasmOperatorComponentID]); err != nil {
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
