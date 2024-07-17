package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func KubePrometheusStandardMonitoring(params metadata.ApplicationParams) metadata.ApplicationStack {
	return metadata.ApplicationStack{
		Components: map[metadata.StackComponentID]metadata.StackComponent{
			metadata.KubePrometheusComponentID: components.KubePrometheusStandardComponent(
				metadata.ComponentParams{
					Version: params.Version,
				},
			),
		},
		StkDepsIdx:  []metadata.StackComponentID{metadata.KubePrometheusComponentID},
		Maintainer:  "github:dipankardas011",
		StackNameID: metadata.KubePrometheusStandardStackID,
	}
}
