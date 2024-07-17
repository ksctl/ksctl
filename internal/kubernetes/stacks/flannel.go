package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	meta "github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func FlannelStandardCNI(params meta.ApplicationParams) meta.ApplicationStack {
	if params.Version == "stable" {
		params.Version = "latest"
	}
	return meta.ApplicationStack{
		Components: map[meta.StackComponentID]meta.StackComponent{
			meta.FlannelComponentID: components.FlannelStandardComponent(
				meta.ComponentParams{
					Version:     params.Version,
					PostInstall: "Flannel (Ver: " + params.Version + ") is a simple and easy way to configure a layer 3 network fabric designed for Kubernetes.",
				},
			),
		},
		StkDepsIdx:  []meta.StackComponentID{meta.FlannelComponentID},
		Maintainer:  "github:dipankardas011",
		StackNameID: meta.FlannelStandardStackID,
	}
}
