package controller

import (
	"context"
	"os"

	"github.com/ksctl/ksctl/api/gen/agent/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(ctx context.Context) (pb.KsctlAgentClient, *grpc.ClientConn, error) {
	ksctlAgentUrl := os.Getenv("KSCTL_AGENT_URL")
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.DialContext(ctx, ksctlAgentUrl, opts...)
	if err != nil {
		return nil, conn, err
	}

	return pb.NewKsctlAgentClient(conn), conn, nil
}

// func appHandler(ctx context.Context, client pb.KsctlAgentClient, apps []applicationv1alpha1.Component, operation pb.ApplicationOperation) error {
// 	if _, ok := helpers.IsContextPresent(ctx, consts.KsctlTestFlagKey); ok {
// 		return nil
// 	}
// 	_apps := make([]*pb.Application, 0)
// 	for _, app := range apps {
// 		_apps = append(_apps, &pb.Application{
// 			AppName: app.AppName,
// 			Version: app.Version,
// 			AppType: func() pb.ApplicationType {
// 				switch app.AppType {
// 				case applicationv1alpha1.TypeCNI:
// 					return pb.ApplicationType_CNI
// 				case applicationv1alpha1.TypeApp:
// 					return pb.ApplicationType_APP
// 				default: // default is app
// 					return pb.ApplicationType_APP
// 				}
// 			}(),
// 		})
// 	}
//
// 	_, err := client.Application(ctx, &pb.ReqApplication{
// 		Operation: operation,
// 		Apps:      _apps,
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
//
// func InstallApps(ctx context.Context, client pb.KsctlAgentClient, apps []applicationv1alpha1.Component) error {
// 	return appHandler(ctx, client, apps, pb.ApplicationOperation_CREATE)
// }
//
// func DeleteApps(ctx context.Context, client pb.KsctlAgentClient, apps []applicationv1alpha1.Component) error {
// 	return appHandler(ctx, client, apps, pb.ApplicationOperation_DELETE)
// }
