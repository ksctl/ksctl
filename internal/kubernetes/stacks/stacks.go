package stacks

import (
	"context"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
)

var appsManifests = map[metadata.StackID]func(metadata.ApplicationParams) (metadata.ApplicationStack, error){
	metadata.ArgocdStandardStackID:          ArgocdStandardCICD,
	metadata.ArgoRolloutsStandardStackID:    ArgoRolloutsStandardCICD,
	metadata.CiliumStandardStackID:          CiliumStandardCNI,
	metadata.FlannelStandardStackID:         FlannelStandardCNI,
	metadata.IstioStandardStackID:           IstioStandardServiceMesh,
	metadata.KubePrometheusStandardStackID:  KubePrometheusStandardMonitoring,
	metadata.KsctlOperatorsID:               KsctlOperatorStackData,
	metadata.SpinKubeProductionStackID:      SpinkubeProductionApp,
	metadata.WasmEdgeKwasmProductionStackID: KwasmWasmedgeProductionApp,
}

func FetchKsctlStack(ctx context.Context, log types.LoggerFactory, stkID string) (func(metadata.ApplicationParams) (metadata.ApplicationStack, error), error) {
	fn, ok := appsManifests[metadata.StackID(stkID)]
	if !ok {
		return nil, ksctlErrors.ErrFailedKsctlComponent.Wrap(
			log.NewError(ctx, "appStack not found", "stkId", string(stkID)),
		)
	}
	return fn, nil
}
