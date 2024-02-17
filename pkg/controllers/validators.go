package controllers

import (
	"errors"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
)

func validationFields(meta resources.Metadata) error {
	log = logger.NewDefaultLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName("ksctl-manager")

	if !helpers.ValidateCloud(meta.Provider) {
		return errors.New("invalid cloud provider")
	}
	if !helpers.ValidateDistro(meta.K8sDistro) {
		return errors.New("invalid kubernetes distro")
	}
	if !helpers.ValidateStorage(meta.StateLocation) {
		return errors.New("invalid storage driver")
	}
	log.Debug("Valid fields from user")
	return nil
}
