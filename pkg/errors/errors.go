// Copyright 2024 ksctl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import "fmt"

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
	ErrPanic

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
	ErrInvalidKsctlClusterAddons

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
	case ErrPanic:
		return "PanicErr"
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
	case ErrInvalidKsctlClusterAddons:
		return "InvalidKsctlClusterAddonsErr"
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

func IsPanic(err error) bool {
	return codeForError(err) == ErrPanic
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

func IsInvalidKsctlClusterAddons(err error) bool {
	return codeForError(err) == ErrInvalidKsctlClusterAddons
}
