package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	meta "github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func FlannelStandardCNI(params meta.ApplicationParams) (meta.ApplicationStack, error) {
	v, err := components.FlannelStandardComponent(
		params.ComponentParams[meta.FlannelComponentID],
	)
	if err != nil {
		return meta.ApplicationStack{}, err
	}

	return meta.ApplicationStack{
		Components: map[meta.StackComponentID]meta.StackComponent{
			meta.FlannelComponentID: v,
		},
		StkDepsIdx:  []meta.StackComponentID{meta.FlannelComponentID},
		Maintainer:  "github:dipankardas011",
		StackNameID: meta.FlannelStandardStackID,
	}, nil
}
