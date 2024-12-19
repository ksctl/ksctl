package pkg_tests_test

import (
	"context"
	"fmt"
	"github.com/ksctl/ksctl/internal/cloudproviders/civo"
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
	fakeClientHA types.CloudFactory
	storeHA      types.StorageFactory

	fakeClientManaged types.CloudFactory
	storeManaged      types.StorageFactory

	fakeClientVars types.CloudFactory
	storeVars      types.StorageFactory

	dir          = filepath.Join(os.TempDir(), "ksctl-civo-pkg-test")
	parentCtx    context.Context
	parentLogger types.LoggerFactory = logger.NewStructuredLogger(-1, os.Stdout)
)

func TestMain(m *testing.M) {
	parentCtx = context.WithValue(context.TODO(), consts.KsctlCustomDirLoc, dir)

	fakeClientVars, _ = civo.NewClient(parentCtx, types.Metadata{
		ClusterName: "demo",
		Region:      "LON1",
		Provider:    consts.CloudCivo,
		IsHA:        true,
	}, parentLogger, &storageTypes.StorageDocument{}, civo.ProvideClient)

	storeVars = localstate.NewClient(parentCtx, parentLogger)
	_ = storeVars.Setup(consts.CloudCivo, "LON1", "demo", consts.ClusterTypeHa)
	_ = storeVars.Connect()

	exitVal := m.Run()

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}
