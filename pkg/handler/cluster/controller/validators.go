package controller

import (
	"github.com/ksctl/ksctl/pkg/config"
	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/validation"
)

func (manager *Controller) ValidateMetadata(c *Client) error {
	meta := c.Metadata
	if _, ok := config.IsContextPresent(manager.ctx, consts.KsctlContextUserID); !ok {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			manager.l.NewError(manager.ctx, "invalid format for context value `USERID`", "Reason", "Make sure the value", "type", "string", "format", `^[\w-]+$`),
		)
	}

	if !validation.ValidateCloud(meta.Provider) {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidCloudProvider,
			manager.l.NewError(
				manager.ctx, "Problem in validation", "cloud", meta.Provider,
			),
		)
	}
	if !validation.ValidateDistro(meta.K8sDistro) {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidBootstrapProvider,
			manager.l.NewError(
				manager.ctx, "Problem in validation", "bootstrap", meta.K8sDistro,
			),
		)
	}

	if err := validation.IsValidName(manager.ctx, manager.l, meta.ClusterName); err != nil {
		return err
	}

	return nil
}
