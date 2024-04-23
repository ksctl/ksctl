package storage

import (
	"context"
	"github.com/ksctl/ksctl/ksctl-components/agent/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
	"os"
)

func NewStorageClient(ctx context.Context, log resources.LoggerFactory, client *resources.KsctlClient) error {
	client.Metadata.StateLocation = consts.StoreK8s
	client.Metadata.LogWritter = helpers.LogWriter
	client.Metadata.LogVerbosity = helpers.LogVerbosity[os.Getenv("LOG_LEVEL")]
	log.Debug("Metadata for Storage", "client.Metadata", client.Metadata)

	return controllers.InitializeStorageFactory(ctx, client)
}

func HandleImport(client *resources.KsctlClient, log resources.LoggerFactory, data *resources.StorageStateExportImport) error {
	err := client.Storage.Import(data)
	if err != nil {
		log.Error("Storage Import failed", "Reason", err)
	}
	return err
}
