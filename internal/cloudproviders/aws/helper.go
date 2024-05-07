package aws

import (
	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
)

func GetInputCredential(storage resources.StorageFactory, meta resources.Metadata) error {
	log = logger.NewDefaultLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName(string(consts.CloudAws))

	log.Print("Enter your AWS ACCESS KEY")
	acesskey, err := helpers.UserInputCredentials(log)
	if err != nil {
		return err
	}

	log.Print("Enter your AWS SECRET KEY")
	acesskeysecret, err := helpers.UserInputCredentials(log)
	if err != nil {
		return err
	}

	apiStore := &types.CredentialsDocument{
		InfraProvider: consts.CloudAws,
		Aws: &types.CredentialsAws{
			AccessKeyId:     acesskey,
			SecretAccessKey: acesskeysecret,
		},
	}

	if err := storage.WriteCredentials(consts.CloudAws, apiStore); err != nil {
		return err
	}

	return nil
}

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

func loadStateHelper(storage resources.StorageFactory) error {
	raw, err := storage.Read()
	if err != nil {
		return log.NewError("Error reading state", "error", err)
	}
	*mainStateDocument = func(x *types.StorageDocument) types.StorageDocument {
		return *x
	}(raw)
	return nil
}
