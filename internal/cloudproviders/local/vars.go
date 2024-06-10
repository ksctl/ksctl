package local

import (
	"context"
	"github.com/ksctl/ksctl/pkg/types"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
)

var (
	mainStateDocument *storageTypes.StorageDocument
	log               types.LoggerFactory
	localCtx          context.Context
)
