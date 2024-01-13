package aws

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/resources"
)

func validationOfArguments(obj *AwsProvider) error {

	if err := isValidRegion(obj, obj.region); err != nil {
		return err
	}

	if err := helpers.IsValidName(obj.clusterName); err != nil {
		return err
	}

	return nil
}

func isValidRegion(obj *AwsProvider, reg string) error {

	ec2client := obj.ec2Client()

	validReg, err := obj.client.ListLocations(ec2client)
	if err != nil {
		return err
	}
	if validReg == nil {
		return fmt.Errorf("no region found")
	}

	return nil
}

// we need to check vm soxe but aws use consts and we have string
// so will check if the string is in the consts

func isValidVMSize(obj *AwsProvider, size string) error {
	validSize, err := obj.client.ListVMTypes(obj.ec2Client())
	if err != nil {
		return err
	}

	for _, valid := range validSize.InstanceTypes {
		constAsString := string(valid.InstanceType)
		if constAsString == size {
			return nil
		}
	}

	return fmt.Errorf("INVALID VM SIZE\nValid options %v\n", validSize)
}

//func isValidK8sVersion(obj *AwsProvider, ver string) error {
//	res, err := obj.client.ListKubernetesVersions()
//	if err != nil {
//		return log.NewError("failed to finish the request: %v", err)
//	}
//
//	log.Debug("Printing", "ListKubernetesVersions", res)
//
//	var vers []string
//	for _, version := range res.Values {
//		vers = append(vers, *version.Version)
//	}
//	for _, valver := range vers {
//		if valver == ver {
//			return nil
//		}
//	}
//	return log.NewError("Invalid k8s version\nValid options: %v\n", vers)
//}

func loadStateHelper(storage resources.StorageFactory) error {
	raw, err := storage.Read()
	if err != nil {
		return log.NewError(err.Error())
	}
	*mainStateDocument = func(x *types.StorageDocument) types.StorageDocument {
		return *x
	}(raw)
	return nil
}
