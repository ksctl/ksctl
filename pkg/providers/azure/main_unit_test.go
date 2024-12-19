package azure

import (
	"context"
	"fmt"
	localstate "github.com/ksctl/ksctl/internal/storage/local"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/types"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
	"os"
	"path/filepath"
	"testing"
)

var (
	fakeClientHA *AzureProvider
	storeHA      types.StorageFactory

	fakeClientManaged *AzureProvider
	storeManaged      types.StorageFactory

	fakeClientVars *AzureProvider
	storeVars      types.StorageFactory
	parentCtx      context.Context
	parentLogger   types.LoggerFactory = logger.NewStructuredLogger(-1, os.Stdout)

	dir = filepath.Join(os.TempDir(), "ksctl-azure-test")
)

func TestMain(m *testing.M) {
	parentCtx = context.WithValue(context.TODO(), consts.KsctlCustomDirLoc, dir)

	fakeClientVars, _ = NewClient(parentCtx, types.Metadata{
		ClusterName: "demo",
		Region:      "fake",
		Provider:    consts.CloudAzure,
		IsHA:        true,
	}, parentLogger, &storageTypes.StorageDocument{}, ProvideClient)

	storeVars = localstate.NewClient(parentCtx, parentLogger)
	_ = storeVars.Setup(consts.CloudAzure, "fake", "demo", consts.ClusterTypeHa)
	_ = storeVars.Connect()

	exitVal := m.Run()

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}
