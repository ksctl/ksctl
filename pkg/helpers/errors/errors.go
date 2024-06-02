package errors

import (
	"fmt"
	"strings"
)

var (
	ErrMissingArgument      = ksctlGlobalErr("MissingArgumentError")
	ErrMissingConfiguration = ksctlGlobalErr("MissingConfigurationError")

	ErrNilCredentials = ksctlGlobalErr("NilCredentialsError")

	ErrTimeOut        = ksctlGlobalErr("TimeoutError")
	ErrSSHFingerprint = ksctlGlobalErr("SSHFingerPrintError")

	ErrInvalidRegion            = ksctlGlobalErr("InvalidRegionError")
	ErrInvalidCloudProvider     = ksctlGlobalErr("InvalidCloudProviderError")
	ErrInvalidBootstrapProvider = ksctlGlobalErr("InvalidBootstrapProvider")
	ErrInvalidStorageProvider   = ksctlGlobalErr("InvalidStorageProviderError")
	ErrInvalidLogger            = ksctlGlobalErr("InvalidLoggerError")
	ErrInvalidResourceName      = ksctlGlobalErr("InvalidResourceNameError")
	ErrInvalidVMSize            = ksctlGlobalErr("InvalidVMSizeError")
	ErrInvalidNoOfControlplane  = ksctlGlobalErr("InvalidNoOfControlplaneError")
	ErrInvalidNoOfLoadbalancer  = ksctlGlobalErr("InvalidNoOfLoadbalancerError")
	ErrInvalidNoOfDatastore     = ksctlGlobalErr("InvalidNoOfDatastoreError")
	ErrInvalidNoOfWorkerplane   = ksctlGlobalErr("InvalidNoOfWorkerplaneError")

	ErrUnknown  = ksctlGlobalErr("UnknownError")
	ErrInternal = ksctlGlobalErr("InternalError")

	ErrFailedCloudResourceQuotaLimitReached = ksctlGlobalErr("FailedCloudResourceQuotaLimitReached")
	ErrFailedInitDatastore                  = ksctlGlobalErr("FailedInitDatastoreError")
	ErrFailedInitControlplane               = ksctlGlobalErr("FailedInitControlplaneError")
	ErrFailedInitWorkerplane                = ksctlGlobalErr("FailedInitWorkerplaneError")
	ErrFailedInitLoadbalancer               = ksctlGlobalErr("FailedInitLoadbalancerError")
	ErrFailedConnectingKubernetesCluster    = ksctlGlobalErr("FailedConnectingKubernetesClusterError")
)

type ksctlGlobalErr string

func (err ksctlGlobalErr) Error() string {
	return string(err)
}

func (err ksctlGlobalErr) Is(target error) bool {
	ts := target.Error()
	es := string(err)
	return ts == es || strings.HasPrefix(ts, es+": ")
}

func (err ksctlGlobalErr) Wrap(inner error) error {
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
	return ksctlGlobalErr(err.code).Is(target)
}
