package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/utils"
)

// TODO: add validation of region, disk size and more

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

func (obj *AzureProvider) setRequiredENV_VAR(storage resources.StorageFactory, ctx context.Context) error {

	envTenant := os.Getenv("AZURE_TENANT_ID")
	envSub := os.Getenv("AZURE_SUBSCRIPTION_ID")
	envClientid := os.Getenv("AZURE_CLIENT_ID")
	envClientsec := os.Getenv("AZURE_CLIENT_SECRET")

	if len(envTenant) != 0 &&
		len(envSub) != 0 &&
		len(envClientid) != 0 &&
		len(envClientsec) != 0 {

		obj.SubscriptionID = envSub
		return nil
	}

	msg := "environment vars not set:"
	if len(envTenant) == 0 {
		msg = msg + " AZURE_TENANT_ID"
	}

	if len(envSub) == 0 {
		msg = msg + " AZURE_SUBSCRIPTION_ID"
	}

	if len(envClientid) == 0 {
		msg = msg + " AZURE_CLIENT_ID"
	}

	if len(envClientsec) == 0 {
		msg = msg + " AZURE_CLIENT_SECRET"
	}

	storage.Logger().Warn(msg)

	tokens, err := utils.GetCred(storage, "azure")
	if err != nil {
		return err
	}

	obj.SubscriptionID = tokens["subscription_id"]

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

func generatePath(flag int, path ...string) string {
	return utils.GetPath(flag, utils.CLOUD_AZURE, path...)
}

func saveStateHelper(storage resources.StorageFactory) error {
	path := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AZURE, clusterType, clusterDirName, STATE_FILE_NAME)
	rawState, err := convertStateToBytes(*azureCloudState)
	if err != nil {
		return err
	}
	return storage.Path(path).Permission(FILE_PERM_CLUSTER_STATE).Save(rawState)
}

func loadStateHelper(storage resources.StorageFactory) error {
	path := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AZURE, clusterType, clusterDirName, STATE_FILE_NAME)
	raw, err := storage.Path(path).Load()
	if err != nil {
		return err
	}

	return convertStateFromBytes(raw)
}

func saveKubeconfigHelper(storage resources.StorageFactory, kubeconfig string) error {
	rawState := []byte(kubeconfig)
	path := utils.GetPath(utils.CLUSTER_PATH, utils.CLOUD_AZURE, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)

	return storage.Path(path).Permission(FILE_PERM_CLUSTER_KUBECONFIG).Save(rawState)
}

func convertStateToBytes(state StateConfiguration) ([]byte, error) {
	return json.Marshal(state)
}

func convertStateFromBytes(raw []byte) error {
	var data *StateConfiguration
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}
	azureCloudState = data
	return nil
}

func printKubeconfig(storage resources.StorageFactory, operation string) {
	env := ""
	storage.Logger().Note("KUBECONFIG env var")
	path := generatePath(utils.CLUSTER_PATH, clusterType, clusterDirName, KUBECONFIG_FILE_NAME)
	switch runtime.GOOS {
	case "windows":
		switch operation {
		case "create":
			env = fmt.Sprintf("$Env:KUBECONFIG=\"%s\"\n", path)
		case "delete":
			env = fmt.Sprintf("$Env:KUBECONFIG=\"\"\n")
		}
	case "linux", "macos":
		switch operation {
		case "create":
			env = fmt.Sprintf("export KUBECONFIG=\"%s\"\n", path)
		case "delete":
			env = "unset KUBECONFIG"
		}
	}
	storage.Logger().Note(env)
}

func validationOfArguments(obj *AzureProvider) error {

	if err := isValidRegion(obj, obj.Region); err != nil {
		return err
	}

	if err := utils.IsValidName(obj.ClusterName); err != nil {
		return err
	}

	return nil
}

// TODO: Add for managed kubernetes version
func isValidK8sVersion(obj *AzureProvider, ver string) error {
	return nil
	//return fmt.Errorf("Invalid k8s version\nValid options: %v\n", vers)
}

func isValidRegion(obj *AzureProvider, reg string) error {
	clientFactory, err := armsubscriptions.NewClientFactory(obj.AzureTokenCred, nil)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	var validReg []string
	pager := clientFactory.NewClient().NewListLocationsPager(obj.SubscriptionID, &armsubscriptions.ClientListLocationsOptions{IncludeExtendedLocations: nil})
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to advance page: %v", err)
		}
		for _, v := range page.Value {
			validReg = append(validReg, *v.Name)
		}
	}

	for _, valid := range validReg {
		if valid == reg {
			return nil
		}
	}
	return fmt.Errorf("INVALID REGION\nValid options: %v\n", validReg)
}

func isValidVMSize(obj *AzureProvider, size string) error {

	clientFactory, err := armcompute.NewClientFactory(obj.SubscriptionID, obj.AzureTokenCred, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	var validSize []string
	pager := clientFactory.NewVirtualMachineSizesClient().NewListPager(obj.Region, nil)
	for pager.More() {

		page, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to advance page: %v", err)
		}
		for _, v := range page.Value {
			validSize = append(validSize, *v.Name)
		}
	}

	for _, valid := range validSize {
		if valid == size {
			return nil
		}
	}

	return fmt.Errorf("INVALID VM SIZE\nValid options %v\n", validSize)
}
