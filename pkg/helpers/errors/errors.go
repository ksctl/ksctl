package errors

import (
	"fmt"
	"strings"
)

var (
	ErrNilCredentials         = ksctlGlobalErr("NilCredentialsErr")
	ErrTimeOut                = ksctlGlobalErr("TimeoutErr")
	ErrContextCancelled       = ksctlGlobalErr("ContextCancelledErr")
	ErrSSHExec                = ksctlGlobalErr("SSHExecErr")
	ErrKubeconfigOperations   = ksctlGlobalErr("KubeconfigOperationsErr")
	ErrUnknown                = ksctlGlobalErr("UnknownErr")
	ErrInternal               = ksctlGlobalErr("InternalErr")
	ErrDuplicateRecords       = ksctlGlobalErr("DuplicateRecordsErr")
	ErrNoMatchingRecordsFound = ksctlGlobalErr("NoMatchingRecordsFoundErr")

	ErrInvalidOperation             = ksctlGlobalErr("InvalidOperationErr")
	ErrInvalidKsctlRole             = ksctlGlobalErr("InvalidKsctlRoleErr")
	ErrInvalidUserInput             = ksctlGlobalErr("InvalidUserInputErr")
	ErrInvalidCloudProvider         = ksctlGlobalErr("InvalidCloudProviderErr")
	ErrInvalidClusterType           = ksctlGlobalErr("InvalidClusterTypeErr")
	ErrInvalidBootstrapProvider     = ksctlGlobalErr("InvalidBootstrapProviderErr")
	ErrInvalidStorageProvider       = ksctlGlobalErr("InvalidStorageProviderErr")
	ErrInvalidResourceName          = ksctlGlobalErr("InvalidResourceNameErr")
	ErrInvalidVersion               = ksctlGlobalErr("InvalidVersion")
	ErrInvalidNoOfControlplane      = ksctlGlobalErr("InvalidNoOfControlplaneErr")
	ErrInvalidNoOfDatastore         = ksctlGlobalErr("InvalidNoOfDatastoreErr")
	ErrInvalidNoOfWorkerplane       = ksctlGlobalErr("InvalidNoOfWorkerplaneErr")
	ErrInvalidKsctlComponentVersion = ksctlGlobalErr("InvalidKsctlComponentVersionErr")

	ErrFailedCloudAccountAuth = ksctlGlobalErr("FailedCloudAccountAuthErr")
	ErrInvalidCloudRegion     = ksctlGlobalErr("InvalidCloudRegionErr")
	ErrInvalidCloudVMSize     = ksctlGlobalErr("InvalidCloudVMSizeErr")

	ErrFailedKsctlComponent              = ksctlGlobalErr("FailedKsctlComponentErr")
	ErrFailedKubernetesClient            = ksctlGlobalErr("FailedKubernetesClientErr")
	ErrFailedHelmClient                  = ksctlGlobalErr("FailedHelmClientErr")
	ErrFailedKsctlClusterOperation       = ksctlGlobalErr("FailedKsctlClusterOperationErr")
	ErrFailedGenerateCertificates        = ksctlGlobalErr("FailedGenerateCertificatesErr")
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
