// Copyright 2024 Ksctl Authors
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

package azure

import (
	"fmt"

	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/types"
)

func generateResourceGroupName(clusterName, clusterType string) string {
	return fmt.Sprintf("ksctl-resgrp-%s-%s", clusterType, clusterName)
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

func validationOfArguments(obj *AzureProvider) error {

	if err := isValidRegion(obj, obj.region); err != nil {
		return err
	}

	return nil
}

func isValidK8sVersion(obj *AzureProvider, ver string) error {
	res, err := obj.client.ListKubernetesVersions()
	if err != nil {
		return err
	}

	log.Debug(azureCtx, "Printing", "ListKubernetesVersions", res)

	var vers []string
	for _, version := range res.Values {
		vers = append(vers, *version.Version)
	}
	for _, valver := range vers {
		if valver == ver {
			return nil
		}
	}

	return ksctlErrors.ErrInvalidVersion.Wrap(
		log.NewError(azureCtx, "invalid k8s version", "ValidManagedK8sVersions", vers),
	)
}

func isValidRegion(obj *AzureProvider, reg string) error {
	validReg, err := obj.client.ListLocations()
	if err != nil {
		return err
	}
	log.Debug(azureCtx, "Printing", "ListLocation", validReg)

	for _, valid := range validReg {
		if valid == reg {
			return nil
		}
	}
	return ksctlErrors.ErrInvalidCloudRegion.Wrap(
		log.NewError(azureCtx, "Invalid region", "Valid options", validReg),
	)
}

func isValidVMSize(obj *AzureProvider, size string) error {

	validSize, err := obj.client.ListVMTypes()
	if err != nil {
		return err
	}
	log.Debug(azureCtx, "Printing", "ListVMType", validSize)

	for _, valid := range validSize {
		if valid == size {
			return nil
		}
	}

	return ksctlErrors.ErrInvalidCloudVMSize.Wrap(
		log.NewError(azureCtx, "Invalid vm size", "Valid options", validSize),
	)
}
