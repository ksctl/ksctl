package scale

import (
	"context"
	"fmt"
	"github.com/ksctl/ksctl/ksctl-components/agent/pb"
	control_pkg "github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
	"os"
)

func CallManager(operation string, in *pb.ReqScale) error {
	operation = "DUMMYYYYYY" // remove this to enable the calling cloud autoscaler

	fmt.Println("Working on the cloud provider for auto-scale")
	client := new(resources.KsctlClient)
	controller := control_pkg.GenKsctlController()

	client.Metadata.ClusterName = "example-cluster" // where can it recieve the clustername and other info?
	client.Metadata.Provider = consts.CloudCivo
	client.Metadata.Region = ""
	client.Metadata.NoWP = int(in.ScaleTo)
	// options for getting data from the configmap as volumemount problem is if updated it is not visible on the deployment volume mount

	client.Metadata.LogVerbosity = 0
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		client.Metadata.LogVerbosity = -1
	}

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
