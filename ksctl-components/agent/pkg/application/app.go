package application

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	ksctlHelpers "github.com/ksctl/ksctl/pkg/helpers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ksctl/ksctl/api/gen/agent/pb"
	control_pkg "github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

func toKsctlControllerCompatableForm(apps []*pb.Application, appType pb.ApplicationType) (_apps []types.KsctlApp, err error) {
	for _, app := range apps {
		if app.AppType == appType {
			_app := types.KsctlApp{}
			if _err := json.Unmarshal(app.AppStackInfo, &_app); _err != nil {
				return nil, _err
			}
			_apps = append(_apps, _app)
		}
	}

	return
}

func Handler(ctx context.Context, log types.LoggerFactory, in *pb.ReqApplication) error {

	client := new(types.KsctlClient)

	client.Metadata.ClusterName = os.Getenv("KSCTL_CLUSTER_NAME")
	client.Metadata.Provider = consts.KsctlCloud(os.Getenv("KSCTL_CLOUD"))
	client.Metadata.K8sDistro = consts.KsctlKubernetes(os.Getenv("KSCTL_K8S_DISTRO"))
	client.Metadata.Region = os.Getenv("KSCTL_REGION")
	client.Metadata.StateLocation = consts.StoreK8s

	v, err := toKsctlControllerCompatableForm(in.Apps, pb.ApplicationType_APP)
	if err != nil {
		stat, estat := status.Newf(
			codes.InvalidArgument,
			fmt.Sprintf("Unable to deserialize App stack info, Context: APP, Reason: %v", err),
		).WithDetails(in)
		if estat != nil {
			return estat
		}
		return stat.Err()
	}
	if len(v) != 0 {
		client.Metadata.Applications = v
	}

	v, err = toKsctlControllerCompatableForm(in.Apps, pb.ApplicationType_CNI)
	if err != nil {
		stat, estat := status.Newf(
			codes.InvalidArgument,
			fmt.Sprintf("Unable to deserialize App stack info, Context: CNI, Reason: %v", err),
		).WithDetails(in)
		if estat != nil {
			return estat
		}
		return stat.Err()
	}
	if len(v) != 0 {
		client.Metadata.CNIPlugin = v[0]
	}

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

	log.Debug(ctx, "Metadata for Application handler", "client.Metadata", client.Metadata)

	if _, ok := ksctlHelpers.IsContextPresent(ctx, consts.KsctlTestFlagKey); ok {
		return nil
	}

	controller, err := control_pkg.NewManagerClusterKubernetes(
		ctx,
		log,
		client,
	)
	if err != nil {
		log.Error("Failed to initialize storage factory", "error", err)

		stat, estat := status.Newf(
			codes.FailedPrecondition,
			fmt.Sprintf("Failed to Initialize storage. Reason: %v", err),
		).WithDetails(in)
		if estat != nil {
			return estat
		}
		return stat.Err()
	}

	var _err error
	switch in.GetOperation() {
	case pb.ApplicationOperation_CREATE:
		log.Debug(ctx, "Application Create")
		_err = controller.ApplicationsAndCni(consts.OperationCreate)
	case pb.ApplicationOperation_DELETE:
		log.Debug(ctx, "Application Delete")
		_err = controller.ApplicationsAndCni(consts.OperationDelete)
	default:
		_err = log.NewError(ctx, "invalid operation")
	}

	if _err != nil {
		stat, estat := status.Newf(
			codes.Internal,
			fmt.Sprintf("Failed to perform operation=%s. Reason: %v", in.GetOperation().String(), _err),
		).WithDetails(in)
		if estat != nil {
			return estat
		}
		return stat.Err()
	}
	return nil
}
