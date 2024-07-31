package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	meta "github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func FlannelStandardCNI(params meta.ApplicationParams) meta.ApplicationStack {

	return meta.ApplicationStack{
		Components: map[meta.StackComponentID]meta.StackComponent{
			meta.FlannelComponentID: components.FlannelStandardComponent(
				params.ComponentParams[meta.FlannelComponentID],
			),
		},
		StkDepsIdx:  []meta.StackComponentID{meta.FlannelComponentID},
		Maintainer:  "github:dipankardas011",
		StackNameID: meta.FlannelStandardStackID,
	}
}
