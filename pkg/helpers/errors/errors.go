package errors

import (
	"fmt"
	"strings"
)

var (
	ErrMissingArgument      = ksctlGlobalErr("MissingArgumentErr")
	ErrMissingConfiguration = ksctlGlobalErr("MissingConfigurationErr")

	ErrNilCredentials = ksctlGlobalErr("NilCredentialsErr")

	ErrTimeOut              = ksctlGlobalErr("TimeoutErr")
	ErrContextCancelled     = ksctlGlobalErr("ContextCancelled")
	ErrSSHExec              = ksctlGlobalErr("SSHExecErr")
	ErrKubeconfigOperations = ksctlGlobalErr("KubeconfigOperationsErr")

	ErrInvalidRegion                = ksctlGlobalErr("InvalidRegionErr")
	ErrInvalidUserInput             = ksctlGlobalErr("InvalidUserInputErr")
	ErrInvalidCloudProvider         = ksctlGlobalErr("InvalidCloudProviderErr")
	ErrInvalidBootstrapProvider     = ksctlGlobalErr("InvalidBootstrapProviderErr")
	ErrInvalidStorageProvider       = ksctlGlobalErr("InvalidStorageProviderErr")
	ErrInvalidLogger                = ksctlGlobalErr("InvalidLoggerErr")
	ErrInvalidResourceName          = ksctlGlobalErr("InvalidResourceNameErr")
	ErrInvalidVMSize                = ksctlGlobalErr("InvalidVMSizeErr")
	ErrInvalidNoOfControlplane      = ksctlGlobalErr("InvalidNoOfControlplaneErr")
	ErrInvalidNoOfLoadbalancer      = ksctlGlobalErr("InvalidNoOfLoadbalancerErr")
	ErrInvalidNoOfDatastore         = ksctlGlobalErr("InvalidNoOfDatastoreErr")
	ErrInvalidNoOfWorkerplane       = ksctlGlobalErr("InvalidNoOfWorkerplaneErr")
	ErrInvalidKsctlComponentVersion = ksctlGlobalErr("InvalidKsctlComponentVersionErr")

	ErrUnknown  = ksctlGlobalErr("UnknownErr")
	ErrInternal = ksctlGlobalErr("InternalErr")

	ErrFailedCloudResourceQuotaLimitReached = ksctlGlobalErr("FailedCloudResourceQuotaLimitReachedErr")
	ErrFailedGenerateCertificates           = ksctlGlobalErr("FailedGenerateCertificatesErr")
	ErrFailedInitDatastore                  = ksctlGlobalErr("FailedInitDatastoreErr")
	ErrFailedInitControlplane               = ksctlGlobalErr("FailedInitControlplaneErr")
	ErrFailedInitWorkerplane                = ksctlGlobalErr("FailedInitWorkerplaneErr")
	ErrFailedInitLoadbalancer               = ksctlGlobalErr("FailedInitLoadbalancerErr")
	ErrFailedConnectingKubernetesCluster    = ksctlGlobalErr("FailedConnectingKubernetesClusterErr")
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
