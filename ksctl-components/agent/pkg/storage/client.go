package storage

import (
	"context"
	"github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
	"os"
)

func NewStorageClient(ctx context.Context, client *resources.KsctlClient) error {
	client.Metadata.StateLocation = consts.StoreK8s
	client.Metadata.LogWritter = os.Stdout
	client.Metadata.LogVerbosity = -1
	return controllers.InitializeStorageFactory(ctx, client)
}

func HandleImport(client *resources.KsctlClient, data *resources.StorageStateExportImport) error {
	err := client.Storage.Import(data)
	return err
}
