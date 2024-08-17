package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func KwasmWasmedgeProductionApp(params metadata.ApplicationParams) (metadata.ApplicationStack, error) {

	kwasmOperatorComponent, err := components.KwasmOperatorComponent(
		params.ComponentParams[metadata.KwasmOperatorComponentID],
	)
	if err != nil {
		return metadata.ApplicationStack{}, err
	}

	wasmedgeKwasmComponent, err := components.KwasmWasmedgeComponent(
		params.ComponentParams[metadata.KwasmRuntimeClassWasmedgeID],
	)
	if err != nil {
		return metadata.ApplicationStack{}, err
	}

	return metadata.ApplicationStack{
		Maintainer:  "github@dipankardas011",
		StackNameID: metadata.WasmEdgeKwasmProductionStackID,
		Components: map[metadata.StackComponentID]metadata.StackComponent{
			metadata.KwasmOperatorComponentID:    kwasmOperatorComponent,
			metadata.KwasmRuntimeClassWasmedgeID: wasmedgeKwasmComponent,
		},
		StkDepsIdx: []metadata.StackComponentID{
			metadata.KwasmOperatorComponentID,
			metadata.KwasmRuntimeClassWasmedgeID,
		},
	}, nil
}
