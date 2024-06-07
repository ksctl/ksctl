package errors

import (
	"fmt"
	"strings"
)

var (
	ErrMissingArgument        = ksctlGlobalErr("MissingArgumentErr")
	ErrMissingConfiguration   = ksctlGlobalErr("MissingConfigurationErr")
	ErrNilCredentials         = ksctlGlobalErr("NilCredentialsErr")
	ErrTimeOut                = ksctlGlobalErr("TimeoutErr")
	ErrContextCancelled       = ksctlGlobalErr("ContextCancelledErr")
	ErrSSHExec                = ksctlGlobalErr("SSHExecErr")
	ErrKubeconfigOperations   = ksctlGlobalErr("KubeconfigOperationsErr")
	ErrUnknown                = ksctlGlobalErr("UnknownErr")
	ErrInternal               = ksctlGlobalErr("InternalErr")
	ErrDuplicateRecords       = ksctlGlobalErr("DuplicateRecordsErr")
	ErrNoMatchingRecordsFound = ksctlGlobalErr("NoMatchingRecordsFoundErr")

	ErrInvalidOperation             = ksctlGlobalErr("InvalidOperation")
	ErrInvalidKsctlRole             = ksctlGlobalErr("InvalidKsctlRole")
	ErrInvalidUserInput             = ksctlGlobalErr("InvalidUserInputErr")
	ErrInvalidCloudProvider         = ksctlGlobalErr("InvalidCloudProviderErr")
	ErrInvalidClusterType           = ksctlGlobalErr("InvalidClusterTypeErr")
	ErrInvalidBootstrapProvider     = ksctlGlobalErr("InvalidBootstrapProviderErr")
	ErrInvalidStorageProvider       = ksctlGlobalErr("InvalidStorageProviderErr")
	ErrInvalidResourceName          = ksctlGlobalErr("InvalidResourceNameErr")
	ErrInvalidVersion               = ksctlGlobalErr("InvalidVersion")
	ErrInvalidNoOfControlplane      = ksctlGlobalErr("InvalidNoOfControlplaneErr")
	ErrInvalidNoOfLoadbalancer      = ksctlGlobalErr("InvalidNoOfLoadbalancerErr")
	ErrInvalidNoOfDatastore         = ksctlGlobalErr("InvalidNoOfDatastoreErr")
	ErrInvalidNoOfWorkerplane       = ksctlGlobalErr("InvalidNoOfWorkerplaneErr")
	ErrInvalidKsctlComponentVersion = ksctlGlobalErr("InvalidKsctlComponentVersionErr")

	ErrFailedCloudResourceQuotaLimitReached = ksctlGlobalErr("FailedCloudResourceQuotaLimitReachedErr")
	ErrFailedCloudAccountAuth               = ksctlGlobalErr("FailedCloudAccountAuthErr")
	ErrInvalidCloudRegion                   = ksctlGlobalErr("InvalidCloudRegionErr")
	ErrInvalidCloudVMSize                   = ksctlGlobalErr("InvalidCloudVMSizeErr")

	ErrFailedKsctlClusterOperation       = ksctlGlobalErr("FailedKsctlClusterOperationErr")
	ErrFailedGenerateCertificates        = ksctlGlobalErr("FailedGenerateCertificatesErr")
	ErrFailedInitDatastore               = ksctlGlobalErr("FailedInitDatastoreErr")
	ErrFailedInitControlplane            = ksctlGlobalErr("FailedInitControlplaneErr")
	ErrFailedInitWorkerplane             = ksctlGlobalErr("FailedInitWorkerplaneErr")
	ErrFailedInitLoadbalancer            = ksctlGlobalErr("FailedInitLoadbalancerErr")
	ErrFailedConnectingKubernetesCluster = ksctlGlobalErr("FailedConnectingKubernetesClusterErr")
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
	return KsctlWrappedError{code: string(err), err: inner}
}

type KsctlWrappedError struct {
	err  error
	code string
}

func (err KsctlWrappedError) Error() string {
	if err.err != nil {
		return fmt.Sprintf("%s: %v", err.code, err.err)
	}
	return err.code
}

func (err KsctlWrappedError) Unwrap() error {
	return err.err
}

func (err KsctlWrappedError) Is(target error) bool {
	return ksctlGlobalErr(err.code).Is(target)
}
