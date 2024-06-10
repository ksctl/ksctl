package external

import (
	"context"

	"github.com/ksctl/ksctl/internal/storage/external/mongodb"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
)

func HandleCreds(ctx context.Context, log types.LoggerFactory, store consts.KsctlStore) (map[string][]byte, error) {
	switch store {
	case consts.StoreLocal, consts.StoreK8s:
		return nil, ksctlErrors.ErrInvalidStorageProvider.Wrap(
			log.NewError(ctx, "these are not external storageProvider"),
		)
	case consts.StoreExtMongo:
		return mongodb.ExportEndpoint()
	default:
		return nil, ksctlErrors.ErrInvalidStorageProvider.Wrap(
			log.NewError(ctx, "invalid storage", "storage", store),
		)
	}
}
