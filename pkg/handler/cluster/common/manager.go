package common

import (
	"context"
	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/pkg/logger"
)

type Controller struct {
	ctx context.Context
	p   *controller.Client
	b   *controller.Controller
}

func NewController(ctx context.Context, log logger.Logger, controllerPayload *controller.Client) (*Controller, error) {
	cc := new(Controller)
	cc.ctx = context.WithValue(ctx, consts.KsctlModuleNameKey, "pkg/handler/cluster/common")
	cc.b = controller.NewBaseController(ctx, log)
	cc.p = controllerPayload

	if err := cc.b.ValidateMetadata(controllerPayload); err != nil {
		return nil, err
	}

	if err := cc.b.InitStorage(controllerPayload); err != nil {
		return nil, err
	}

	if err := cc.b.StartPoller(); err != nil {
		return nil, err
	}

	return cc, nil
}
