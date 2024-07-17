package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func KsctlOperatorStackData(params metadata.ApplicationParams) metadata.ApplicationStack {
	if params.Version == "latest" {
		params.Version = "main"
	}
	return metadata.ApplicationStack{
		StackNameID: metadata.KsctlOperatorsID,
		Maintainer:  "github@dipankardas011",
		Components: map[metadata.StackComponentID]metadata.StackComponent{
			metadata.KsctlApplicationComponentID: components.KsctlApplicationComponent(
				metadata.ComponentParams{
					Version: params.Version,
					PostInstall: `
As the controller and the crd are installed just need to apply application to be installed
`,
				},
			),
		},
		StkDepsIdx: []metadata.StackComponentID{
			metadata.KsctlApplicationComponentID,
		},
	}
}
