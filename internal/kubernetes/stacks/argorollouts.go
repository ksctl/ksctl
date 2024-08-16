package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func ArgoRolloutsStandardCICD(params metadata.ApplicationParams) (metadata.ApplicationStack, error) {

	return metadata.ApplicationStack{
		Components: map[metadata.StackComponentID]metadata.StackComponent{
			metadata.ArgorolloutsComponentID: components.ArgoRolloutsStandardComponent(
				params.ComponentParams[metadata.ArgorolloutsComponentID],
			),
		},

		StkDepsIdx:  []metadata.StackComponentID{metadata.ArgorolloutsComponentID},
		StackNameID: metadata.ArgoRolloutsStandardStackID,
		Maintainer:  "github@dipankardas011",
	}, nil
}
