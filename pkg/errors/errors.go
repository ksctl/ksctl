package errors

import "fmt"

// Usage
//err := NewError(ErrNilCredentials)
//
//// Wrap an existing error
//err = WrapError(ErrTimeOut, someError)
//
//// Create a formatted error
//err = WrapErrorf(ErrInvalidUserInput, "invalid input: %s", input)
//
//// Check error types
//if IsTimeout(err) {
//// handle timeout
//}

type ErrorCode int

const (
	ErrNilCredentials ErrorCode = iota
	ErrTimeOut
	ErrContextCancelled
	ErrSSHExec
	ErrKubeconfigOperations
	ErrUnknown
	ErrInternal
	ErrDuplicateRecords
	ErrNoMatchingRecordsFound

	ErrInvalidOperation
	ErrInvalidKsctlRole
	ErrInvalidUserInput
	ErrInvalidCloudProvider
	ErrInvalidClusterType
	ErrInvalidBootstrapProvider
	ErrInvalidStorageProvider
	ErrInvalidResourceName
	ErrInvalidVersion
	ErrInvalidNoOfControlplane
	ErrInvalidNoOfDatastore
	ErrInvalidNoOfWorkerplane
	ErrInvalidKsctlComponentVersion

	ErrFailedCloudAccountAuth
	ErrInvalidCloudRegion
	ErrInvalidCloudVMSize

	ErrFailedKsctlComponent
	ErrFailedKubernetesClient
	ErrFailedHelmClient
	ErrFailedKsctlClusterOperation
	ErrFailedGenerateCertificates
	ErrFailedConnectingKubernetesCluster
)

// KsctlError is the error type for ksctl errors
type KsctlError struct {
	err  error
	code ErrorCode
}

// Error returns the error message
func (e KsctlError) Error() string {
	if e.err == nil {
		return e.errorCodeToString()
	}
	return fmt.Sprintf("%s: %v", e.errorCodeToString(), e.err)
}

// errorCodeToString converts an ErrorCode to its string representation
func (e KsctlError) errorCodeToString() string {
	switch e.code {
	case ErrNilCredentials:
		return "NilCredentialsErr"
	case ErrTimeOut:
		return "TimeoutErr"
	case ErrContextCancelled:
		return "ContextCancelledErr"
	case ErrSSHExec:
		return "SSHExecErr"
	case ErrKubeconfigOperations:
		return "KubeconfigOperationsErr"
	case ErrUnknown:
		return "UnknownErr"
	case ErrInternal:
		return "InternalErr"
	case ErrDuplicateRecords:
		return "DuplicateRecordsErr"
	case ErrNoMatchingRecordsFound:
		return "NoMatchingRecordsFoundErr"
	case ErrInvalidOperation:
		return "InvalidOperationErr"
	case ErrInvalidKsctlRole:
		return "InvalidKsctlRoleErr"
	case ErrInvalidUserInput:
		return "InvalidUserInputErr"
	case ErrInvalidCloudProvider:
		return "InvalidCloudProviderErr"
	case ErrInvalidClusterType:
		return "InvalidClusterTypeErr"
	case ErrInvalidBootstrapProvider:
		return "InvalidBootstrapProviderErr"
	case ErrInvalidStorageProvider:
		return "InvalidStorageProviderErr"
	case ErrInvalidResourceName:
		return "InvalidResourceNameErr"
	case ErrInvalidVersion:
		return "InvalidVersion"
	case ErrInvalidNoOfControlplane:
		return "InvalidNoOfControlplaneErr"
	case ErrInvalidNoOfDatastore:
		return "InvalidNoOfDatastoreErr"
	case ErrInvalidNoOfWorkerplane:
		return "InvalidNoOfWorkerplaneErr"
	case ErrInvalidKsctlComponentVersion:
		return "InvalidKsctlComponentVersionErr"
	case ErrFailedCloudAccountAuth:
		return "FailedCloudAccountAuthErr"
	case ErrInvalidCloudRegion:
		return "InvalidCloudRegionErr"
	case ErrInvalidCloudVMSize:
		return "InvalidCloudVMSizeErr"
	case ErrFailedKsctlComponent:
		return "FailedKsctlComponentErr"
	case ErrFailedKubernetesClient:
		return "FailedKubernetesClientErr"
	case ErrFailedHelmClient:
		return "FailedHelmClientErr"
	case ErrFailedKsctlClusterOperation:
		return "FailedKsctlClusterOperationErr"
	case ErrFailedGenerateCertificates:
		return "FailedGenerateCertificatesErr"
	case ErrFailedConnectingKubernetesCluster:
		return "FailedConnectingKubernetesClusterErr"
	default:
		return "UnknownError"
	}
}

// WrapError wraps an error with an error code
func WrapError(c ErrorCode, e error) error {
	return KsctlError{err: e, code: c}
}

// WrapErrorf wraps an error with an error code and a format string
func WrapErrorf(c ErrorCode, format string, args ...interface{}) error {
	return KsctlError{err: fmt.Errorf(format, args...), code: c}
}

// NewError creates a new error with just an error code
func NewError(c ErrorCode) error {
	return KsctlError{code: c}
}

func codeForError(err error) ErrorCode {
	if v, ok := err.(KsctlError); ok {
		return v.code
	}
	return -1
}

func IsNilCredentials(err error) bool {
	return codeForError(err) == ErrNilCredentials
}

func IsTimeout(err error) bool {
	return codeForError(err) == ErrTimeOut
}

func IsContextCancelled(err error) bool {
	return codeForError(err) == ErrContextCancelled
}

func IsSSHExec(err error) bool {
	return codeForError(err) == ErrSSHExec
}

func IsKubeconfigOperations(err error) bool {
	return codeForError(err) == ErrKubeconfigOperations
}

func IsUnknown(err error) bool {
	return codeForError(err) == ErrUnknown
}

func IsInternal(err error) bool {
	return codeForError(err) == ErrInternal
}

func IsDuplicateRecords(err error) bool {
	return codeForError(err) == ErrDuplicateRecords
}

func IsNoMatchingRecordsFound(err error) bool {
	return codeForError(err) == ErrNoMatchingRecordsFound
}

func IsInvalidOperation(err error) bool {
	return codeForError(err) == ErrInvalidOperation
}

func IsInvalidKsctlRole(err error) bool {
	return codeForError(err) == ErrInvalidKsctlRole
}

func IsInvalidUserInput(err error) bool {
	return codeForError(err) == ErrInvalidUserInput
}

func IsInvalidCloudProvider(err error) bool {
	return codeForError(err) == ErrInvalidCloudProvider
}

func IsInvalidClusterType(err error) bool {
	return codeForError(err) == ErrInvalidClusterType
}

func IsInvalidBootstrapProvider(err error) bool {
	return codeForError(err) == ErrInvalidBootstrapProvider
}

func IsInvalidStorageProvider(err error) bool {
	return codeForError(err) == ErrInvalidStorageProvider
}

func IsInvalidResourceName(err error) bool {
	return codeForError(err) == ErrInvalidResourceName
}

func IsInvalidVersion(err error) bool {
	return codeForError(err) == ErrInvalidVersion
}

func IsInvalidNoOfControlplane(err error) bool {
	return codeForError(err) == ErrInvalidNoOfControlplane
}

func IsInvalidNoOfDatastore(err error) bool {
	return codeForError(err) == ErrInvalidNoOfDatastore
}

func IsInvalidNoOfWorkerplane(err error) bool {
	return codeForError(err) == ErrInvalidNoOfWorkerplane
}

func IsInvalidKsctlComponentVersion(err error) bool {
	return codeForError(err) == ErrInvalidKsctlComponentVersion
}

func IsFailedCloudAccountAuth(err error) bool {
	return codeForError(err) == ErrFailedCloudAccountAuth
}

func IsInvalidCloudRegion(err error) bool {
	return codeForError(err) == ErrInvalidCloudRegion
}

func IsInvalidCloudVMSize(err error) bool {
	return codeForError(err) == ErrInvalidCloudVMSize
}

func IsFailedKsctlComponent(err error) bool {
	return codeForError(err) == ErrFailedKsctlComponent
}

func IsFailedKubernetesClient(err error) bool {
	return codeForError(err) == ErrFailedKubernetesClient
}

func IsFailedHelmClient(err error) bool {
	return codeForError(err) == ErrFailedHelmClient
}

func IsFailedKsctlClusterOperation(err error) bool {
	return codeForError(err) == ErrFailedKsctlClusterOperation
}

func IsFailedGenerateCertificates(err error) bool {
	return codeForError(err) == ErrFailedGenerateCertificates
}

func IsFailedConnectingKubernetesCluster(err error) bool {
	return codeForError(err) == ErrFailedConnectingKubernetesCluster
}

func IsKsctlError(err error) bool {
	_, ok := err.(KsctlError)
	return ok
}