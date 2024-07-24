package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	applicationv1alpha1 "github.com/ksctl/ksctl/ksctl-components/operators/application/api/v1alpha1"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"

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

func appHandler(
	ctx context.Context,
	client pb.KsctlAgentClient,
	apps applicationv1alpha1.StackSpec,
	operation pb.ApplicationOperation,
) error {

	if _, ok := helpers.IsContextPresent(ctx, consts.KsctlTestFlagKey); ok {
		return nil
	}
	_apps := make([]*pb.Application, 0)

	for _, stack := range apps.Stacks {
		ksctlApplicationData := types.KsctlApp{
			StackName: stack.StackId,
		}

		_overrides := stack.Overrides.Raw
		if _overrides != nil {
			ksctlApplicationData.Overrides = make(map[string]map[string]any)
			fmt.Printf("Overrides: %#v\n", _overrides)
			if err := json.Unmarshal(_overrides, &ksctlApplicationData.Overrides); err != nil {
				log.Error("Unmarshal", "Reason", err)
				return err
			}
		}
		var appType pb.ApplicationType
		switch stack.AppType {
		case applicationv1alpha1.TypeApp:
			appType = pb.ApplicationType_APP
		case applicationv1alpha1.TypeCNI:
			appType = pb.ApplicationType_CNI
		default:
			appType = pb.ApplicationType_APP
		}

		raw_app, err := json.Marshal(ksctlApplicationData)
		if err != nil {
			return err
		}
		_apps = append(_apps, &pb.Application{
			AppType:      appType,
			AppStackInfo: raw_app,
		})
	}

	_, err := client.Application(ctx, &pb.ReqApplication{
		Operation: operation,
		Apps:      _apps,
	})
	if err != nil {
		return err
	}
	return nil
}

func InstallApps(
	ctx context.Context,
	client pb.KsctlAgentClient,
	apps applicationv1alpha1.StackSpec,
) error {
	return appHandler(
		ctx,
		client,
		apps,
		pb.ApplicationOperation_CREATE)
}

func DeleteApps(
	ctx context.Context,
	client pb.KsctlAgentClient,
	apps applicationv1alpha1.StackSpec,
) error {
	return appHandler(
		ctx,
		client,
		apps,
		pb.ApplicationOperation_DELETE)
}
