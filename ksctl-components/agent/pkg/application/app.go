package application

import (
	"context"
	"os"

	"github.com/ksctl/ksctl/api/gen/agent/pb"
	control_pkg "github.com/ksctl/ksctl/pkg/controllers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
)

func toKsctlControllerCompatableForm(app []*pb.Application, appType pb.ApplicationType) (_apps []string) {
	for _, app := range app {
		if app.AppType == appType {
			_app := ""
			if len(app.Version) == 0 {
				_app = app.AppName
			} else {
				_app = app.AppName + "@" + app.Version
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

	if v := toKsctlControllerCompatableForm(in.Apps, pb.ApplicationType_APP); len(v) != 0 {
		client.Metadata.Applications = v
	}
	if v := toKsctlControllerCompatableForm(in.Apps, pb.ApplicationType_CNI); len(v) != 0 {
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

	controller, err := control_pkg.GenKsctlController(
		ctx,
		log,
		client,
	)
	if err != nil {
		log.Error(ctx, "Failed to initialize storage factory", "error", err)
		return err
	}

	switch in.Operation {
	case pb.ApplicationOperation_CREATE:
		log.Debug(ctx, "Application Create")
		return controller.Applications(consts.OperationCreate)
	case pb.ApplicationOperation_DELETE:
		log.Debug(ctx, "Application Delete")
		return controller.Applications(consts.OperationDelete)
	default:
		return log.NewError(ctx, "invalid operation")
	}
}
