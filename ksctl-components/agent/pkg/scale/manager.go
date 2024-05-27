package scale

import (
	"context"
	"fmt"

	"github.com/ksctl/ksctl/api/gen/agent/pb"

	"os"

	control_pkg "github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

func CallManager(ctx context.Context, log types.LoggerFactory, in *pb.ReqScale) error {

	// TODO: make the manager to not use kuberneteVer as a parameter instead it to be handled by the scaling thing
	// Reason: we can update the ver without changing the env; by just changing the state along and it should be the single source of truth!
	client := new(types.KsctlClient)

	client.Metadata.ClusterName = os.Getenv("KSCTL_CLUSTER_NAME")
	client.Metadata.Provider = consts.KsctlCloud(os.Getenv("KSCTL_CLOUD"))
	client.Metadata.K8sDistro = consts.KsctlKubernetes(os.Getenv("KSCTL_K8S_DISTRO"))
	client.Metadata.Region = os.Getenv("KSCTL_REGION")
	client.Metadata.NoWP = int(in.DesiredNoOfWP)
	client.Metadata.WorkerPlaneNodeType = in.NodeSizeOfWP
	client.Metadata.StateLocation = consts.StoreK8s
	client.Metadata.IsHA = func() bool {
		var _v bool
		switch os.Getenv("KSCTL_CLUSTER_IS_HA") {
		case "true":
			return true
		case "false":
			return false
		}
		return _v
	}()

	controller, err := control_pkg.NewManagerClusterSelfManaged(
		ctx,
		log,
		client,
	)
	if err != nil {
		return err
	}

	log.Debug(ctx, "Metadata for Application handler", "client.Metadata", client.Metadata)

	switch in.Operation {
	case pb.ScaleOperation_SCALE_UP:
		return controller.AddWorkerPlaneNodes()
	case pb.ScaleOperation_SCALE_DOWN:
		return controller.DelWorkerPlaneNodes()
	default:
		return fmt.Errorf("invalid operation")
	}
}
