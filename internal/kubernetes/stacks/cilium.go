package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	meta "github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func CiliumStandardCNI(params meta.ApplicationParams) meta.ApplicationStack {
	return meta.ApplicationStack{
		Components: map[meta.StackComponentID]meta.StackComponent{
			meta.CiliumComponentID: components.CiliumStandardComponent(
				params.ComponentParams[meta.CiliumComponentID],
			),
		},

		StkDepsIdx:  []meta.StackComponentID{meta.CiliumComponentID},
		StackNameID: meta.CiliumStandardStackID,
		Maintainer:  "github@dipankardas011",
	}
}
