package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func KubePrometheusStandardMonitoring(params metadata.ApplicationParams) (metadata.ApplicationStack, error) {
	return metadata.ApplicationStack{
		Components: map[metadata.StackComponentID]metadata.StackComponent{
			metadata.KubePrometheusComponentID: components.KubePrometheusStandardComponent(
				params.ComponentParams[metadata.KubePrometheusComponentID],
			),
		},
		StkDepsIdx:  []metadata.StackComponentID{metadata.KubePrometheusComponentID},
		Maintainer:  "github:dipankardas011",
		StackNameID: metadata.KubePrometheusStandardStackID,
	}, nil
}
