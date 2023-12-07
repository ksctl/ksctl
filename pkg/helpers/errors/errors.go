package errors

import (
	"fmt"
	"strings"
)

var (
	ErrMissingArgument      = globalError("MissingArgumentError")
	ErrMissingConfiguration = globalError("MissingConfigurationError")

	ErrNilCredentials = globalError("NilCredentialsError")

	ErrTimeOut        = globalError("TimeoutError")
	ErrSSHFingerprint = globalError("SSHFingerPrintError")

	ErrInvalidRegion           = globalError("InvalidRegionError")
	ErrInvalidCloudProvider    = globalError("InvalidCloudProviderError")
	ErrInvalidDistribution     = globalError("InvalidDistributionError")
	ErrInvalidStorage          = globalError("InvalidStorageError")
	ErrInvalidLogger           = globalError("InvalidLoggerError")
	ErrInvalidResourceName     = globalError("InvalidResourceNameError")
	ErrInvalidVMSize           = globalError("InvalidVMSizeError")
	ErrInvalidNoOfControlplane = globalError("InvalidNoOfControlplaneError")
	ErrInvalidNoOfLoadbalancer = globalError("InvalidNoOfLoadbalancerError")
	ErrInvalidNoOfDatastore    = globalError("InvalidNoOfDatastoreError")
	ErrInvalidNoOfWorkerplane  = globalError("InvalidNoOfWorkerplaneError")

	ErrUnknown  = globalError("UnknownError")
	ErrInternal = globalError("InternalError")

	ErrFailedInitDatastore               = globalError("FailedInitDatastoreError")
	ErrFailedInitControlplane            = globalError("FailedInitControlplaneError")
	ErrFailedInitWorkerplane             = globalError("FailedInitWorkerplaneError")
	ErrFailedInitLoadbalancer            = globalError("FailedInitLoadbalancerError")
	ErrFailedConnectingKubernetesCluster = globalError("FailedConnectingKubernetesClusterError")
)

type globalError string

func (err globalError) Error() string {
	return string(err)
}

func (err globalError) Is(target error) bool {
	ts := target.Error()
	es := string(err)
	return ts == es || strings.HasPrefix(ts, es+": ")
}

func (err globalError) Wrap(inner error) error {
	return wrapError{code: string(err), err: inner}
}

type wrapError struct {
	err  error
	code string
}

func (err wrapError) Error() string {
	if err.err != nil {
		return fmt.Sprintf("%s: %v", err.code, err.err)
	}
	return err.code
}

func (err wrapError) Unwrap() error {
	return err.err
}

func (err wrapError) Is(target error) bool {
	return globalError(err.code).Is(target)
}
