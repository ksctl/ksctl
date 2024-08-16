package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func KsctlOperatorStackData(params metadata.ApplicationParams) (metadata.ApplicationStack, error) {

	return metadata.ApplicationStack{
		StackNameID: metadata.KsctlOperatorsID,
		Maintainer:  "github@dipankardas011",
		Components: map[metadata.StackComponentID]metadata.StackComponent{
			metadata.KsctlApplicationComponentID: components.KsctlApplicationComponent(
				params.ComponentParams[metadata.KsctlApplicationComponentID],
			),
		},
		StkDepsIdx: []metadata.StackComponentID{
			metadata.KsctlApplicationComponentID,
		},
	}, nil
}
