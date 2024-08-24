package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func ArgoRolloutsStandardCICD(params metadata.ApplicationParams) (metadata.ApplicationStack, error) {
	v, err := components.ArgoRolloutsStandardComponent(
		params.ComponentParams[metadata.ArgorolloutsComponentID],
	)
	if err != nil {
		return metadata.ApplicationStack{}, err
	}

	return metadata.ApplicationStack{
		Components: map[metadata.StackComponentID]metadata.StackComponent{
			metadata.ArgorolloutsComponentID: v,
		},

		StkDepsIdx:  []metadata.StackComponentID{metadata.ArgorolloutsComponentID},
		StackNameID: metadata.ArgoRolloutsStandardStackID,
		Maintainer:  "github@dipankardas011",
	}, nil
}
