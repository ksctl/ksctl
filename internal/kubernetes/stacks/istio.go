package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	meta "github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func IstioStandardServiceMesh(params meta.ApplicationParams) meta.ApplicationStack {
	return meta.ApplicationStack{
		Components: map[meta.StackComponentID]meta.StackComponent{
			meta.IstioComponentID: components.IstioStandardComponent(
				meta.ComponentParams{
					Version: params.Version,
				},
			),
		},
		StkDepsIdx: []meta.StackComponentID{
			meta.IstioComponentID,
		},
		Maintainer:  "github:dipankardas011",
		StackNameID: meta.IstioStandardStackID,
	}
}
