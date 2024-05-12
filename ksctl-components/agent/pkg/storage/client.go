package storage

import (
	"context"

	"github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

func NewStorageClient(ctx context.Context, log types.LoggerFactory, data *types.StorageStateExportImport, client *types.KsctlClient) error {
	client.Metadata.StateLocation = consts.StoreK8s
	log.Debug(ctx, "Metadata for Storage", "client.Metadata", client.Metadata)

	_, err := controllers.GenKsctlController(ctx, log, client)
	if err != nil {
		return err
	}
	err = client.Storage.Import(data)
	if err != nil {
		return err
	}
	return err
}
