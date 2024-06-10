package aws

import (
	"context"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
)

var (
	mainStateDocument *storageTypes.StorageDocument
	clusterType       consts.KsctlClusterType
	log               types.LoggerFactory
	awsCtx            context.Context
)
