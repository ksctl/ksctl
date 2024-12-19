package azure

import (
	"context"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
)

var (
	mainStateDocument *storageTypes.StorageDocument
	clusterType       consts.KsctlClusterType // it stores the ha or managed
	azureCtx          context.Context
	log               types.LoggerFactory
)
