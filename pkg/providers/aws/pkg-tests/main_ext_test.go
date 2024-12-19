package pkg_tests_test

import (
	"context"
	"fmt"
	"github.com/ksctl/ksctl/internal/cloudproviders/aws"
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
	parentCtx         context.Context
	fakeClientVars    types.CloudFactory
	storeVars         types.StorageFactory

	parentLogger types.LoggerFactory = logger.NewStructuredLogger(-1, os.Stdout)

	dir = filepath.Join(os.TempDir(), "ksctl-aws-pkg-test")
)

func TestMain(m *testing.M) {

	parentCtx = context.WithValue(
		context.TODO(),
		consts.KsctlCustomDirLoc,
		dir)

	fakeClientVars, _ = aws.NewClient(parentCtx, types.Metadata{
		ClusterName: "demo",
		Region:      "fake-region",
		Provider:    consts.CloudAws,
		IsHA:        true,
	}, parentLogger, &storageTypes.StorageDocument{}, aws.ProvideClient)

	storeVars = localstate.NewClient(parentCtx, parentLogger)
	_ = storeVars.Setup(consts.CloudAws, "fake-region", "demo", consts.ClusterTypeHa)
	_ = storeVars.Connect()

	exitVal := m.Run()

	fmt.Println("Cleanup..")
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}
