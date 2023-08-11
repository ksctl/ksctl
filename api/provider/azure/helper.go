package azure

import (
	"context"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
	"os"
)

func GetInputCredential(storage resources.StorageFactory) error {

	storage.Logger().Print("Enter your SUBSCRIPTION ID")
	skey, err := utils.UserInputCredentials(storage.Logger())
	if err != nil {
		return err
	}

	storage.Logger().Print("Enter your TENANT ID")
	tid, err := utils.UserInputCredentials(storage.Logger())
	if err != nil {
		return err
	}

	storage.Logger().Print("Enter your CLIENT ID")
	cid, err := utils.UserInputCredentials(storage.Logger())
	if err != nil {
		return err
	}

	storage.Logger().Print("Enter your CLIENT SECRET")
	cs, err := utils.UserInputCredentials(storage.Logger())
	if err != nil {
		return err
	}

	apiStore := Credential{
		SubscriptionID: skey,
		TenantID:       tid,
		ClientID:       cid,
		ClientSecret:   cs,
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

	if err := utils.SaveCred(storage, apiStore, utils.CLOUD_AZURE); err != nil {
		return err
	}

	return nil
}

func setRequiredENV_VAR(storage resources.StorageFactory, ctx context.Context, cred *AzureProvider) error {

	env_tenant := os.Getenv("AZURE_TENANT_ID")
	env_sub := os.Getenv("AZURE_SUBSCRIPTION_ID")
	env_clientid := os.Getenv("AZURE_CLIENT_ID")
	env_clientsec := os.Getenv("AZURE_CLIENT_SECRET")

	if len(env_tenant) != 0 &&
		len(env_sub) != 0 &&
		len(env_clientid) != 0 &&
		len(env_clientsec) != 0 {

		cred.SubscriptionID = env_sub
		return nil
	}

	msg := "environment vars not set:"
	if len(env_tenant) == 0 {
		msg = msg + " AZURE_TENANT_ID"
	}

	if len(env_sub) == 0 {
		msg = msg + " AZURE_SUBSCRIPTION_ID"
	}

	if len(env_clientid) == 0 {
		msg = msg + " AZURE_CLIENT_ID"
	}

	if len(env_clientsec) == 0 {
		msg = msg + " AZURE_CLIENT_SECRET"
	}

	storage.Logger().Warn(msg)

	tokens, err := utils.GetCred(storage, "azure")
	if err != nil {
		return err
	}

	cred.SubscriptionID = tokens["subscription_id"]

	err = os.Setenv("AZURE_SUBSCRIPTION_ID", tokens["subscription_id"])
	if err != nil {
		return err
	}

	err = os.Setenv("AZURE_TENANT_ID", tokens["tenant_id"])
	if err != nil {
		return err
	}

	err = os.Setenv("AZURE_CLIENT_ID", tokens["client_id"])
	if err != nil {
		return err
	}

	err = os.Setenv("AZURE_CLIENT_SECRET", tokens["client_secret"])
	if err != nil {
		return err
	}
	return nil
}
