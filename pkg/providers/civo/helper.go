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

package civo

import (
	"os"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
)

func fetchAPIKey(storage types.StorageFactory) (string, error) {

	civoToken := os.Getenv("CIVO_TOKEN")
	if civoToken != "" {
		return civoToken, nil
	}
	log.Debug(civoCtx, "environment vars not set: `CIVO_TOKEN`")

	credentials, err := storage.ReadCredentials(consts.CloudCivo)
	if err != nil {
		return "", err
	}
	if credentials.Civo == nil {
		return "", ksctlErrors.ErrNilCredentials.Wrap(
			log.NewError(civoCtx, "no credentials was found"),
		)
	}
	return credentials.Civo.Token, nil
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

func getValidK8sVersionClient(obj *CivoProvider) ([]string, error) {
	vers, err := obj.client.ListAvailableKubernetesVersions()
	if err != nil {
		return nil, err
	}
	log.Debug(civoCtx, "Printing", "ListAvailableKubernetesVersions", vers)
	var val []string
	for _, ver := range vers {
		if ver.ClusterType == string(consts.K8sK3s) {
			val = append(val, ver.Label)
		}
	}
	return val, nil
}

func getValidRegionsClient(obj *CivoProvider) ([]string, error) {
	regions, err := obj.client.ListRegions()
	if err != nil {
		return nil, err
	}
	log.Debug(civoCtx, "Printing", "ListRegions", regions)
	var val []string
	for _, region := range regions {
		val = append(val, region.Code)
	}
	return val, nil
}

func getValidVMSizesClient(obj *CivoProvider) ([]string, error) {
	nodeSizes, err := obj.client.ListInstanceSizes()
	if err != nil {
		return nil, err
	}
	log.Debug(civoCtx, "Printing", "ListInstanceSizes", nodeSizes)
	var val []string
	for _, region := range nodeSizes {
		val = append(val, region.Name)
	}
	return val, nil
}

func validationOfArguments(obj *CivoProvider) error {

	if err := isValidRegion(obj, obj.region); err != nil {
		return err
	}

	return nil
}

func isValidK8sVersion(obj *CivoProvider, ver string) error {
	valver, err := getValidK8sVersionClient(obj)
	if err != nil {
		return err
	}
	for _, vver := range valver {
		if vver == ver {
			return nil
		}
	}
	return ksctlErrors.ErrInvalidVersion.Wrap(
		log.NewError(civoCtx, "invalid k8s version", "ValidManagedK8sVersions", valver),
	)
}

func isValidRegion(obj *CivoProvider, reg string) error {
	validFromClient, err := getValidRegionsClient(obj)
	if err != nil {
		return err
	}
	for _, region := range validFromClient {
		if region == reg {
			return nil
		}
	}
	return ksctlErrors.ErrInvalidCloudRegion.Wrap(
		log.NewError(civoCtx, "invalid region", "ValidRegion", validFromClient),
	)
}

func isValidVMSize(obj *CivoProvider, size string) error {
	validFromClient, err := getValidVMSizesClient(obj)
	if err != nil {
		return err
	}
	for _, nodeSize := range validFromClient {
		if size == nodeSize {
			return nil
		}
	}
	return ksctlErrors.ErrInvalidCloudVMSize.Wrap(
		log.NewError(civoCtx, "invalid Virtual Machine size", "ValidVMSize", validFromClient),
	)
}
