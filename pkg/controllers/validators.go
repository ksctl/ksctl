package controllers

import (
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
)

func (manager *managerInfo) validationFields(meta types.Metadata) error {

	if _, ok := helpers.IsContextPresent(controllerCtx, consts.KsctlContextUserID); !ok {
		return ksctlErrors.ErrInvalidUserInput.Wrap(
			manager.log.NewError(controllerCtx, "invalid format for context value `USERID`", "Reason", "Make sure the value", "type", "string", "format", `^[\w-]+$`),
		)
	}

	if !helpers.ValidateCloud(meta.Provider) {
		return ksctlErrors.ErrInvalidCloudProvider.Wrap(
			manager.log.NewError(
				controllerCtx, "Problem in validation", "cloud", meta.Provider,
			),
		)
	}
	if !helpers.ValidateDistro(meta.K8sDistro) {
		return ksctlErrors.ErrInvalidBootstrapProvider.Wrap(
			manager.log.NewError(
				controllerCtx, "Problem in validation", "bootstrap", meta.K8sDistro,
			),
		)
	}

	return nil
}
