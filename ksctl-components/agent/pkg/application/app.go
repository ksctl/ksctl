package application

import (
	"context"
	"os"

	"github.com/ksctl/ksctl/api/gen/agent/pb"
	"github.com/ksctl/ksctl/ksctl-components/agent/pkg/helpers"
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

func Handler(log types.LoggerFactory, in *pb.ReqApplication) error {

	client := new(types.KsctlClient)
	controller := control_pkg.GenKsctlController()

	client.Metadata.ClusterName = os.Getenv("KSCTL_CLUSTER_NAME")
	client.Metadata.Provider = consts.KsctlCloud(os.Getenv("KSCTL_CLOUD"))
	client.Metadata.K8sDistro = consts.KsctlKubernetes(os.Getenv("KSCTL_K8S_DISTRO"))
	client.Metadata.Region = os.Getenv("KSCTL_REGION")
	client.Metadata.LogVerbosity = helpers.LogVerbosity[os.Getenv("LOG_LEVEL")]
	client.Metadata.StateLocation = consts.StoreK8s
	client.Metadata.LogWritter = helpers.LogWriter

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

	log.Debug("Metadata for Application handler", "client.Metadata", client.Metadata)

	if err := control_pkg.InitializeStorageFactory(context.WithValue(context.Background(), "USERID", "ksctl-agent"), client); err != nil {
		log.Error("Failed to initialize storage factory", "error", err)
		return err
	}

	switch in.Operation {
	case pb.ApplicationOperation_CREATE:
		log.Debug("Application Create")
		return controller.Applications(client, consts.OperationCreate)
	case pb.ApplicationOperation_DELETE:
		log.Debug("Application Delete")
		return controller.Applications(client, consts.OperationDelete)
	default:
		return log.NewError("invalid operation")
	}
}
