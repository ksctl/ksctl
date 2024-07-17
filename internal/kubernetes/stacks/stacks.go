package stacks

import "github.com/ksctl/ksctl/internal/kubernetes/metadata"

var (
	AppsManifests = map[metadata.StackID]func(metadata.ApplicationParams) metadata.ApplicationStack{
		metadata.ArgocdStandardStackID:         ArgocdStandardCICD,
		metadata.ArgoRolloutsStandardStackID:   ArgoRolloutsStandardCICD,
		metadata.CiliumStandardStackID:         CiliumStandardCNI,
		metadata.FlannelStandardStackID:        FlannelStandardCNI,
		metadata.IstioStandardStackID:          IstioStandardServiceMesh,
		metadata.KubePrometheusStandardStackID: KubePrometheusStandardMonitoring,
		metadata.KsctlOperatorsID:              KsctlOperatorStackData,
	}
)
