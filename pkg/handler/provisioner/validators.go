package provisioner

import (
	"github.com/ksctl/ksctl/pkg/config"
	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/validation"
)

func (manager *managerInfo) validationFields(meta Metadata) error {

	if _, ok := config.IsContextPresent(controllerCtx, consts.KsctlContextUserID); !ok {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidUserInput,
			manager.log.NewError(controllerCtx, "invalid format for context value `USERID`", "Reason", "Make sure the value", "type", "string", "format", `^[\w-]+$`),
		)
	}

	if !validation.ValidateCloud(meta.Provider) {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidCloudProvider,
			manager.log.NewError(
				controllerCtx, "Problem in validation", "cloud", meta.Provider,
			),
		)
	}
	if !validation.ValidateDistro(meta.K8sDistro) {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidBootstrapProvider,
			manager.log.NewError(
				controllerCtx, "Problem in validation", "bootstrap", meta.K8sDistro,
			),
		)
	}

	return nil
}
