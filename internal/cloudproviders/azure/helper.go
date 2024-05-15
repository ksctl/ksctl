package azure

import (
	"fmt"

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
	return log.NewError(azureCtx, "Invalid k8s version", "Valid options", vers)
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
	return log.NewError(azureCtx, "Invalid region", "Valid options", validReg)
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

	return log.NewError(azureCtx, "Invalid vm size", "Valid options", validSize)
}
