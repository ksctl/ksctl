package scale

import (
	"context"
	"fmt"

	"github.com/ksctl/ksctl/api/gen/agent/pb"
	"github.com/ksctl/ksctl/ksctl-components/agent/pkg/helpers"

	"os"

	control_pkg "github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

func CallManager(in *pb.ReqScale) error {

	// TODO: make the manager to not use kuberneteVer as a parameter instead it to be handled by the scaling thing
	// Reason: we can update the ver without changing the env; by just changing the state along and it should be the single source of truth!
	client := new(resources.KsctlClient)
	controller := control_pkg.GenKsctlController()

	client.Metadata.ClusterName = os.Getenv("KSCTL_CLUSTER_NAME")
	client.Metadata.Provider = consts.KsctlCloud(os.Getenv("KSCTL_CLOUD"))
	client.Metadata.K8sDistro = consts.KsctlKubernetes(os.Getenv("KSCTL_K8S_DISTRO"))
	client.Metadata.Region = os.Getenv("KSCTL_REGION")
	client.Metadata.NoWP = int(in.DesiredNoOfWP)
	client.Metadata.WorkerPlaneNodeType = in.NodeSizeOfWP // FIXME: we should have consistent nodeSize not a variable; then we can drop this
	client.Metadata.LogVerbosity = helpers.LogVerbosity[os.Getenv("LOG_LEVEL")]
	client.Metadata.StateLocation = consts.StoreK8s
	client.Metadata.LogWritter = os.Stdout
	client.Metadata.IsHA = true

	if err := control_pkg.InitializeStorageFactory(context.WithValue(context.Background(), "USERID", "ksctl-agent"), client); err != nil {
		return err
	}

	switch in.Operation {
	case pb.ScaleOperation_SCALE_UP:
		return controller.AddWorkerPlaneNode(client)
	case pb.ScaleOperation_SCALE_DOWN:
		return controller.DelWorkerPlaneNode(client)
	default:
		return fmt.Errorf("invalid operation")
	}
}
