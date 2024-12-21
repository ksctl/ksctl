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

package aws

import (
	"github.com/ksctl/ksctl/pkg/helpers"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
)

func validationOfArguments(obj *AwsProvider) error {

	if err := isValidRegion(obj); err != nil {
		return err
	}

	if err := helpers.IsValidName(awsCtx, log, obj.clusterName); err != nil {
		return err
	}

	return nil
}

func isValidRegion(obj *AwsProvider) error {

	validReg, err := obj.client.ListLocations()
	if err != nil {
		return err
	}

	for _, reg := range validReg {
		if reg == obj.region {
			return nil
		}
	}

	return ksctlErrors.ErrInvalidCloudRegion.Wrap(
		log.NewError(awsCtx, "region not found", "validRegions", validReg),
	)
}

func isValidVMSize(obj *AwsProvider, size string) error {
	validSize, err := obj.client.ListVMTypes()
	if err != nil {
		return err
	}

	for _, valid := range validSize.InstanceTypes {
		constAsString := string(valid.InstanceType)
		if constAsString == size {
			return nil
		}
	}

	return ksctlErrors.ErrInvalidCloudVMSize.Wrap(
		log.NewError(awsCtx, "invalid vm size", "Valid options", validSize),
	)
}

func loadStateHelper(storage types.StorageFactory) error {
	raw, err := storage.Read()
	if err != nil {
		return err
	}
	*mainStateDocument = func(x *storageTypes.StorageDocument) storageTypes.StorageDocument {
		return *x
	}(raw)
	return nil
}

func isValidK8sVersion(obj *AwsProvider, version string) error {
	validVersions, err := obj.client.ListK8sVersions(awsCtx)
	if err != nil {
		return err
	}

	for _, ver := range validVersions {
		if ver == version {
			return nil
		}
	}

	return ksctlErrors.ErrInvalidVersion.Wrap(
		log.NewError(awsCtx, "invalid k8s version", "validVersions", validVersions),
	)
}
