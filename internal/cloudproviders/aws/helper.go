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

	if validReg == nil { // FIXME: do we actually need this?
		return log.NewError(awsCtx, "no region found")
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

	return log.NewError(awsCtx, "invalid vm size", "Valid options", validSize)
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
