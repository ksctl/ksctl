package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func KwasmProductionApp(params metadata.ApplicationParams) (metadata.ApplicationStack, error) {

	kwasmOperatorComponent, err := components.KwasmOperatorComponent(
		params.ComponentParams[metadata.KwasmOperatorComponentID],
	)
	if err != nil {
		return metadata.ApplicationStack{}, err
	}

	wasmedgeKwasmComponent, err := components.KwasmComponent(
		params.ComponentParams[metadata.KwasmRuntimeClassID],
	)
	if err != nil {
		return metadata.ApplicationStack{}, err
	}

	return metadata.ApplicationStack{
		Maintainer:  "github@dipankardas011",
		StackNameID: metadata.KwasmProductionStackID,
		Components: map[metadata.StackComponentID]metadata.StackComponent{
			metadata.KwasmOperatorComponentID: kwasmOperatorComponent,
			metadata.KwasmRuntimeClassID:      wasmedgeKwasmComponent,
		},
		StkDepsIdx: []metadata.StackComponentID{
			metadata.KwasmOperatorComponentID,
			metadata.KwasmRuntimeClassID,
		},
	}, nil
}
