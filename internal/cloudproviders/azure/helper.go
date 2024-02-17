package azure

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/logger"
	"github.com/ksctl/ksctl/pkg/resources"
)

func GenerateResourceGroupName(clusterName, clusterType string) string {
	return fmt.Sprintf("ksctl-resgrp-%s-%s", clusterType, clusterName)
}

func GetInputCredential(storage resources.StorageFactory, meta resources.Metadata) error {
	log = logger.NewDefaultLogger(meta.LogVerbosity, meta.LogWritter)
	log.SetPackageName(string(consts.CloudAws))

	log.Print("Enter your SUBSCRIPTION ID")
	skey, err := helpers.UserInputCredentials(log)
	if err != nil {
		return err
	}

	log.Print("Enter your TENANT ID")
	tid, err := helpers.UserInputCredentials(log)
	if err != nil {
		return err
	}

	log.Print("Enter your CLIENT ID")
	cid, err := helpers.UserInputCredentials(log)
	if err != nil {
		return err
	}

	log.Print("Enter your CLIENT SECRET")
	cs, err := helpers.UserInputCredentials(log)
	if err != nil {
		return err
	}

	apiStore := &types.CredentialsDocument{
		InfraProvider: consts.CloudAzure,
		Azure: &types.CredentialsAzure{
			SubscriptionID: skey,
			TenantID:       tid,
			ClientID:       cid,
			ClientSecret:   cs,
		},
	}

	// FIXME: add ping pong for validation of credentials
	//if err = os.Setenv("AZURE_SUBSCRIPTION_ID", skey); err != nil {
	//	return err
	//}
	//
	//if err = os.Setenv("AZURE_TENANT_ID", tid); err != nil {
	//	return err
	//}
	//
	//if err = os.Setenv("AZURE_CLIENT_ID", cid); err != nil {
	//	return err
	//}
	//
	//if err = os.Setenv("AZURE_CLIENT_SECRET", cs); err != nil {
	//	return err
	//}
	// ADD SOME PING method to validate credentials

	if err := storage.WriteCredentials(consts.CloudAzure, apiStore); err != nil {
		return err
	}

	return nil
}

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

func validationOfArguments(obj *AzureProvider) error {

	if err := isValidRegion(obj, obj.region); err != nil {
		return err
	}

	return nil
}

func isValidK8sVersion(obj *AzureProvider, ver string) error {
	res, err := obj.client.ListKubernetesVersions()
	if err != nil {
		return log.NewError("failed to finish the request: %v", err)
	}

	log.Debug("Printing", "ListKubernetesVersions", res)

	var vers []string
	for _, version := range res.Values {
		vers = append(vers, *version.Version)
	}
	for _, valver := range vers {
		if valver == ver {
			return nil
		}
	}
	return log.NewError("Invalid k8s version\nValid options: %v\n", vers)
}

func isValidRegion(obj *AzureProvider, reg string) error {
	validReg, err := obj.client.ListLocations()
	if err != nil {
		return err
	}
	log.Debug("Printing", "ListLocation", validReg)

	for _, valid := range validReg {
		if valid == reg {
			return nil
		}
	}
	return log.NewError("INVALID REGION\nValid options: %v\n", validReg)
}

func isValidVMSize(obj *AzureProvider, size string) error {

	validSize, err := obj.client.ListVMTypes()
	if err != nil {
		return err
	}
	log.Debug("Printing", "ListVMType", validSize)

	for _, valid := range validSize {
		if valid == size {
			return nil
		}
	}

	return log.NewError("INVALID VM SIZE\nValid options %v\n", validSize)
}
