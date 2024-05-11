package aws

import (
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/types"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
)

func validationOfArguments(obj *AwsProvider) error {

	if err := isValidRegion(obj); err != nil {
		return err
	}

	if err := helpers.IsValidName(obj.clusterName); err != nil {
		return err
	}

	return nil
}

func isValidRegion(obj *AwsProvider) error {

	validReg, err := obj.client.ListLocations()
	if err != nil {
		return err
	}
	if validReg == nil {
		return log.NewError("no region found")
	}

	return nil
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

	return log.NewError("INVALID VM SIZE\nValid options %v\n", validSize)
}

func loadStateHelper(storage types.StorageFactory) error {
	raw, err := storage.Read()
	if err != nil {
		return log.NewError("Error reading state", "error", err)
	}
	*mainStateDocument = func(x *storageTypes.StorageDocument) storageTypes.StorageDocument {
		return *x
	}(raw)
	return nil
}
