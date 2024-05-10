package controllers

import (
	"errors"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/resources"
)

func validationFields(meta resources.Metadata) error {

	if !helpers.ValidateCloud(meta.Provider) {
		return errors.New("invalid cloud provider")
	}
	if !helpers.ValidateDistro(meta.K8sDistro) {
		return errors.New("invalid kubernetes distro")
	}
	if !helpers.ValidateStorage(meta.StateLocation) {
		return errors.New("invalid storage driver")
	}
	return nil
}
