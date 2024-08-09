package controller

import (
	"context"
	"encoding/json"
	"os"

	"github.com/gookit/goutil/dump"
	applicationv1alpha1 "github.com/ksctl/ksctl/ksctl-components/operators/application/api/v1alpha1"
	"github.com/ksctl/ksctl/pkg/types"

	"github.com/ksctl/ksctl/api/gen/agent/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type KsctlAgent interface {
	InstallApps(apps applicationv1alpha1.StackSpec) error
	UninstallApps(apps applicationv1alpha1.StackSpec) error
	Close() error
}

type KsctlAgentClient struct {
	ctx           context.Context
	ksctlAgentUrl string
	client        pb.KsctlAgentClient
	conn          *grpc.ClientConn
}

func NewKsctlAgentClient(ctx context.Context) (*KsctlAgentClient, error) {
	ksctlAgentUrl := os.Getenv("KSCTL_AGENT_URL")
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.DialContext(ctx, ksctlAgentUrl, opts...)
	if err != nil {
		return nil, err
	}

	return &KsctlAgentClient{
		ctx:           ctx,
		ksctlAgentUrl: ksctlAgentUrl,
		client:        pb.NewKsctlAgentClient(conn),
		conn:          conn,
	}, nil
}

func (c *KsctlAgentClient) Close() error {
	return c.conn.Close()
}

func convertToClientType(apps applicationv1alpha1.StackSpec) ([]*pb.Application, error) {
	_apps := make([]*pb.Application, 0)

	for _, stack := range apps.Stacks {
		ksctlApplicationData := types.KsctlApp{
			StackName: stack.StackId,
		}
		if stack.Overrides != nil {
			_overrides := stack.Overrides.Raw
			if _overrides != nil {
				ksctlApplicationData.Overrides = make(map[string]map[string]any)
				if err := json.Unmarshal(_overrides, &ksctlApplicationData.Overrides); err != nil {
					log.Error("Unmarshal", "Reason", err)
					return nil, err
				}
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
			return nil, err
		}
		_apps = append(_apps, &pb.Application{
			AppType:      appType,
			AppStackInfo: raw_app,
		})
	}
	return _apps, nil
}

func (c *KsctlAgentClient) InstallApps(apps applicationv1alpha1.StackSpec) error {
	_apps, err := convertToClientType(apps)
	if err != nil {
		return err
	}

	_, err = c.client.Application(c.ctx, &pb.ReqApplication{
		Operation: pb.ApplicationOperation_CREATE,
		Apps:      _apps,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *KsctlAgentClient) UninstallApps(apps applicationv1alpha1.StackSpec) error {
	_apps, err := convertToClientType(apps)
	if err != nil {
		return err
	}

	_, err = c.client.Application(c.ctx, &pb.ReqApplication{
		Operation: pb.ApplicationOperation_DELETE,
		Apps:      _apps,
	})
	if err != nil {
		return err
	}
	return nil
}

type KsctlAgentClientMock struct{}

func NewKsctlAgentClientTesting(_ context.Context) (*KsctlAgentClientMock, error) {
	return &KsctlAgentClientMock{}, nil
}

func (c *KsctlAgentClientMock) Close() error {
	return nil
}

func (c *KsctlAgentClientMock) InstallApps(apps applicationv1alpha1.StackSpec) error {
	_apps, err := convertToClientType(apps)
	if err != nil {
		return err
	}

	dump.Println(_apps)

	return nil
}

func (c *KsctlAgentClientMock) UninstallApps(apps applicationv1alpha1.StackSpec) error {
	_apps, err := convertToClientType(apps)
	if err != nil {
		return err
	}

	dump.Println(_apps)

	return nil
}
