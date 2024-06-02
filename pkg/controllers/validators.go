package controllers

import (
	"github.com/ksctl/ksctl/pkg/helpers"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
)

func (manager *managerInfo) validationFields(meta types.Metadata) error {

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
