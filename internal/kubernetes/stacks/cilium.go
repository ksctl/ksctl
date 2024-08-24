package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	meta "github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func CiliumStandardCNI(params meta.ApplicationParams) (meta.ApplicationStack, error) {
	v, err := components.CiliumStandardComponent(
		params.ComponentParams[meta.CiliumComponentID],
	)
	if err != nil {
		return meta.ApplicationStack{}, err
	}

	return meta.ApplicationStack{
		Components: map[meta.StackComponentID]meta.StackComponent{
			meta.CiliumComponentID: v,
		},

		StkDepsIdx:  []meta.StackComponentID{meta.CiliumComponentID},
		StackNameID: meta.CiliumStandardStackID,
		Maintainer:  "github@dipankardas011",
	}, nil
}
