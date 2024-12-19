package civo

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
	fakeClientHA *CivoProvider
	storeHA      types.StorageFactory

	fakeClientManaged *CivoProvider
	storeManaged      types.StorageFactory

	fakeClientVars *CivoProvider
	storeVars      types.StorageFactory

	dir          = filepath.Join(os.TempDir(), "ksctl-civo-test")
	parentCtx    context.Context
	parentLogger types.LoggerFactory = logger.NewStructuredLogger(-1, os.Stdout)
)

func TestMain(m *testing.M) {
	parentCtx = context.WithValue(context.TODO(), consts.KsctlCustomDirLoc, dir)

	fakeClientVars, _ = NewClient(parentCtx, types.Metadata{
		ClusterName: "demo",
		Region:      "LON1",
		Provider:    consts.CloudCivo,
		IsHA:        true,
	}, parentLogger, &storageTypes.StorageDocument{}, ProvideClient)

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
