package scale

import (
	"context"
	"fmt"
	control_pkg "github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
	"os"
)

func CallManager(operation string) error {
	operation = "DUMMYYYYYY" // remove this to enable the calling cloud autoscaler

	fmt.Println("Working on the cloud provider for auto-scale")
	client := new(resources.KsctlClient)
	controller := control_pkg.GenKsctlController()

	client.Metadata.ClusterName = "example-cluster"
	client.Metadata.StateLocation = consts.StoreLocal
	client.Metadata.K8sDistro = consts.K8sK3s

	client.Metadata.LogVerbosity = 0
	client.Metadata.LogWritter = os.Stdout

	if err := control_pkg.InitializeStorageFactory(context.WithValue(context.Background(), "USERID", "ksctl-agent"), client); err != nil {
		return err
	}

	switch operation {
	case "scaleup":
		return controller.AddWorkerPlaneNode(client)
	case "scaledown":
		return controller.DelWorkerPlaneNode(client)
	}
	return nil
}
