package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"

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

func validationOfArguments(name, region string) error {

	if err := isValidRegion(region); err != nil {
		return err
	}

	if err := utils.IsValidName(name); err != nil {
		return err
	}

	return nil
}

// TODO: Add them
func isValidK8sVersion(obj *AzureProvider, ver string) error {
	return nil
	//return fmt.Errorf("Invalid k8s version\nValid options: %v\n", vers)
}

// TODO: use the azure client for validation
func isValidRegion(reg string) error {
	validRegions := []string{"eastus", "eastus2", "southcentralus", "westus2", "westus3", "australiaeast", "southeastasia", "northeurope", "swedencentral", "uksouth", "westeurope", "centralus", "southafricanorth", "centralindia", "eastasia", "japaneast", "koreacentral", "canadacentral", "francecentral", "germanywestcentral", "norwayeast", "switzerlandnorth", "uaenorth", "brazilsouth", "centraluseuap", "eastus2euap", "qatarcentral", "centralusstage", "eastusstage", "eastus2stage", "northcentralusstage", "southcentralusstage", "westusstage", "westus2stage", "asia", "asiapacific", "australia", "brazil", "canada", "europe", "france", "germany", "global", "india", "japan", "korea", "norway", "singapore", "southafrica", "switzerland", "uae", "uk", "unitedstates", "unitedstateseuap", "eastasiastage", "southeastasiastage", "brazilus", "eastusstg", "northcentralus", "westus", "jioindiawest", "devfabric", "westcentralus", "southafricawest", "australiacentral", "australiacentral2", "australiasoutheast", "japanwest", "jioindiacentral", "koreasouth", "southindia", "westindia", "canadaeast", "francesouth", "germanynorth", "norwaywest", "switzerlandwest", "ukwest", "uaecentral", "brazilsoutheast"}

	for _, validRegion := range validRegions {
		if validRegion == reg {
			return nil
		}
	}
	return fmt.Errorf("INVALID REGION\nValid options: %v\n", validRegions)
}

// TODO: use the azure client for validation
func isValidVMSize(size string) error {
	validDisks := []string{
		"Standard_B1ms", "Standard_B1s", "Standard_B2ms", "Standard_B1ls", "Standard_B2s", "Standard_B4ms", "Standard_B8ms", "Standard_B12ms", "Standard_B16ms", "Standard_B20ms", "Standard_E2_v4", "Standard_E4_v4", "Standard_E8_v4", "Standard_E16_v4", "Standard_E20_v4", "Standard_E32_v4", "Standard_E2d_v4", "Standard_E4d_v4", "Standard_E8d_v4", "Standard_E16d_v4", "Standard_E20d_v4", "Standard_E32d_v4", "Standard_E2s_v4", "Standard_E4-2s_v4", "Standard_E4s_v4", "Standard_E8-2s_v4", "Standard_E8-4s_v4", "Standard_E8s_v4", "Standard_E16-4s_v4", "Standard_E16-8s_v4", "Standard_E16s_v4", "Standard_E20s_v4", "Standard_E32-8s_v4", "Standard_E32-16s_v4", "Standard_E32s_v4", "Standard_E2ds_v4", "Standard_E4-2ds_v4", "Standard_E4ds_v4", "Standard_E8-2ds_v4", "Standard_E8-4ds_v4", "Standard_E8ds_v4", "Standard_E16-4ds_v4", "Standard_E16-8ds_v4", "Standard_E16ds_v4", "Standard_E20ds_v4", "Standard_E32-8ds_v4", "Standard_E32-16ds_v4", "Standard_E32ds_v4", "Standard_D2d_v4", "Standard_D4d_v4", "Standard_D8d_v4", "Standard_D16d_v4", "Standard_D32d_v4", "Standard_D48d_v4", "Standard_D64d_v4", "Standard_D2_v4", "Standard_D4_v4", "Standard_D8_v4", "Standard_D16_v4", "Standard_D32_v4", "Standard_D48_v4", "Standard_D64_v4", "Standard_D2ds_v4", "Standard_D4ds_v4", "Standard_D8ds_v4", "Standard_D16ds_v4", "Standard_D32ds_v4", "Standard_D48ds_v4", "Standard_D64ds_v4", "Standard_D2s_v4", "Standard_D4s_v4", "Standard_D8s_v4", "Standard_D16s_v4", "Standard_D32s_v4", "Standard_D48s_v4", "Standard_D64s_v4", "Standard_D1_v2", "Standard_D2_v2", "Standard_D3_v2", "Standard_D4_v2", "Standard_D5_v2", "Standard_D11_v2", "Standard_D12_v2", "Standard_D13_v2", "Standard_D14_v2", "Standard_D15_v2", "Standard_D2_v2_Promo", "Standard_D3_v2_Promo", "Standard_D4_v2_Promo", "Standard_D5_v2_Promo", "Standard_D11_v2_Promo", "Standard_D12_v2_Promo", "Standard_D13_v2_Promo", "Standard_D14_v2_Promo", "Standard_F1", "Standard_F2", "Standard_F4", "Standard_F8", "Standard_F16", "Standard_DS1_v2", "Standard_DS2_v2", "Standard_DS3_v2", "Standard_DS4_v2", "Standard_DS5_v2", "Standard_DS11-1_v2", "Standard_DS11_v2", "Standard_DS12-1_v2", "Standard_DS12-2_v2", "Standard_DS12_v2", "Standard_DS13-2_v2", "Standard_DS13-4_v2", "Standard_DS13_v2", "Standard_DS14-4_v2", "Standard_DS14-8_v2", "Standard_DS14_v2", "Standard_DS15_v2", "Standard_DS2_v2_Promo", "Standard_DS3_v2_Promo", "Standard_DS4_v2_Promo", "Standard_DS5_v2_Promo", "Standard_DS11_v2_Promo", "Standard_DS12_v2_Promo", "Standard_DS13_v2_Promo", "Standard_DS14_v2_Promo", "Standard_F1s", "Standard_F2s", "Standard_F4s", "Standard_F8s", "Standard_F16s", "Standard_A1_v2", "Standard_A2m_v2", "Standard_A2_v2", "Standard_A4m_v2", "Standard_A4_v2", "Standard_A8m_v2", "Standard_A8_v2", "Standard_D2_v3", "Standard_D4_v3", "Standard_D8_v3", "Standard_D16_v3", "Standard_D32_v3", "Standard_D48_v3", "Standard_D64_v3", "Standard_D2s_v3", "Standard_D4s_v3", "Standard_D8s_v3", "Standard_D16s_v3", "Standard_D32s_v3", "Standard_D48s_v3", "Standard_D64s_v3", "Standard_E2_v3", "Standard_E4_v3", "Standard_E8_v3", "Standard_E16_v3", "Standard_E20_v3", "Standard_E32_v3", "Standard_E2s_v3", "Standard_E4-2s_v3", "Standard_E4s_v3", "Standard_E8-2s_v3", "Standard_E8-4s_v3", "Standard_E8s_v3", "Standard_E16-4s_v3", "Standard_E16-8s_v3", "Standard_E16s_v3", "Standard_E20s_v3", "Standard_E32-8s_v3", "Standard_E32-16s_v3", "Standard_E32s_v3", "Standard_F2s_v2", "Standard_F4s_v2", "Standard_F8s_v2", "Standard_F16s_v2", "Standard_F32s_v2", "Standard_F48s_v2", "Standard_F64s_v2", "Standard_F72s_v2", "Standard_D2a_v4", "Standard_D4a_v4", "Standard_D8a_v4", "Standard_D16a_v4", "Standard_D32a_v4", "Standard_D48a_v4", "Standard_D64a_v4", "Standard_D96a_v4", "Standard_D2as_v4", "Standard_D4as_v4", "Standard_D8as_v4", "Standard_D16as_v4", "Standard_D32as_v4", "Standard_D48as_v4", "Standard_D64as_v4", "Standard_D96as_v4", "Standard_E2a_v4", "Standard_E4a_v4", "Standard_E8a_v4", "Standard_E16a_v4", "Standard_E20a_v4", "Standard_E32a_v4", "Standard_E48a_v4", "Standard_E64a_v4", "Standard_E96a_v4", "Standard_E2as_v4", "Standard_E4-2as_v4", "Standard_E4as_v4", "Standard_E8-2as_v4", "Standard_E8-4as_v4", "Standard_E8as_v4", "Standard_E16-4as_v4", "Standard_E16-8as_v4", "Standard_E16as_v4", "Standard_E20as_v4", "Standard_E32-8as_v4", "Standard_E32-16as_v4", "Standard_E32as_v4", "Standard_E48as_v4", "Standard_E64-16as_v4", "Standard_E64-32as_v4", "Standard_E64as_v4", "Standard_E96-24as_v4", "Standard_E96-48as_v4", "Standard_E96as_v4", "Standard_D2as_v5", "Standard_D4as_v5", "Standard_D8as_v5", "Standard_D16as_v5", "Standard_D32as_v5", "Standard_D48as_v5", "Standard_D64as_v5", "Standard_D96as_v5", "Standard_E2as_v5", "Standard_E4-2as_v5", "Standard_E4as_v5", "Standard_E8-2as_v5", "Standard_E8-4as_v5", "Standard_E8as_v5", "Standard_E16-4as_v5", "Standard_E16-8as_v5", "Standard_E16as_v5", "Standard_E20as_v5", "Standard_E32-8as_v5", "Standard_E32-16as_v5", "Standard_E32as_v5", "Standard_E48as_v5", "Standard_E64-16as_v5", "Standard_E64-32as_v5", "Standard_E64as_v5", "Standard_E96-24as_v5", "Standard_E96-48as_v5", "Standard_E96as_v5", "Standard_E112ias_v5", "Standard_D2ads_v5", "Standard_D4ads_v5", "Standard_D8ads_v5", "Standard_D16ads_v5", "Standard_D32ads_v5", "Standard_D48ads_v5", "Standard_D64ads_v5", "Standard_D96ads_v5", "Standard_E2ads_v5", "Standard_E4-2ads_v5", "Standard_E4ads_v5", "Standard_E8-2ads_v5", "Standard_E8-4ads_v5", "Standard_E8ads_v5", "Standard_E16-4ads_v5", "Standard_E16-8ads_v5", "Standard_E16ads_v5", "Standard_E20ads_v5", "Standard_E32-8ads_v5", "Standard_E32-16ads_v5", "Standard_E32ads_v5", "Standard_E48ads_v5", "Standard_E64-16ads_v5", "Standard_E64-32ads_v5", "Standard_E64ads_v5", "Standard_E96-24ads_v5", "Standard_E96-48ads_v5", "Standard_E96ads_v5", "Standard_E112iads_v5", "Standard_E48_v4", "Standard_E64_v4", "Standard_E48d_v4", "Standard_E64d_v4", "Standard_E48s_v4", "Standard_E64-16s_v4", "Standard_E64-32s_v4", "Standard_E64s_v4", "Standard_E80is_v4", "Standard_E48ds_v4", "Standard_E64-16ds_v4", "Standard_E64-32ds_v4", "Standard_E64ds_v4", "Standard_E80ids_v4", "Standard_E48_v3", "Standard_E64_v3", "Standard_E48s_v3", "Standard_E64-16s_v3", "Standard_E64-32s_v3", "Standard_E64s_v3", "Standard_A0", "Standard_A1", "Standard_A2", "Standard_A3", "Standard_A5", "Standard_A4", "Standard_A6", "Standard_A7", "Basic_A0", "Basic_A1", "Basic_A2", "Basic_A3", "Basic_A4", "Standard_L8as_v3", "Standard_L16as_v3", "Standard_L32as_v3", "Standard_L48as_v3", "Standard_L64as_v3", "Standard_L80as_v3", "Standard_NC4as_T4_v3", "Standard_NC8as_T4_v3", "Standard_NC16as_T4_v3", "Standard_NC64as_T4_v3", "Standard_M64", "Standard_M64m", "Standard_M128", "Standard_M128m", "Standard_M8-2ms", "Standard_M8-4ms", "Standard_M8ms", "Standard_M16-4ms", "Standard_M16-8ms", "Standard_M16ms", "Standard_M32-8ms", "Standard_M32-16ms", "Standard_M32ls", "Standard_M32ms", "Standard_M32ts", "Standard_M64-16ms", "Standard_M64-32ms", "Standard_M64ls", "Standard_M64ms", "Standard_M64s", "Standard_M128-32ms", "Standard_M128-64ms", "Standard_M128ms", "Standard_M128s", "Standard_M32ms_v2", "Standard_M64ms_v2", "Standard_M64s_v2", "Standard_M128ms_v2", "Standard_M128s_v2", "Standard_M192ims_v2", "Standard_M192is_v2", "Standard_M32dms_v2", "Standard_M64dms_v2", "Standard_M64ds_v2", "Standard_M128dms_v2", "Standard_M128ds_v2", "Standard_M192idms_v2", "Standard_M192ids_v2", "Standard_D1", "Standard_D2", "Standard_D3", "Standard_D4", "Standard_D11", "Standard_D12", "Standard_D13", "Standard_D14", "Standard_DS1", "Standard_DS2", "Standard_DS3", "Standard_DS4", "Standard_DS11", "Standard_DS12", "Standard_DS13", "Standard_DS14", "Standard_DC8_v2", "Standard_DC1s_v2", "Standard_DC2s_v2", "Standard_DC4s_v2", "Standard_D2ds_v5", "Standard_D4ds_v5", "Standard_D8ds_v5", "Standard_D16ds_v5", "Standard_D32ds_v5", "Standard_D48ds_v5", "Standard_D64ds_v5", "Standard_D96ds_v5", "Standard_D2d_v5", "Standard_D4d_v5", "Standard_D8d_v5", "Standard_D16d_v5", "Standard_D32d_v5", "Standard_D48d_v5", "Standard_D64d_v5", "Standard_D96d_v5", "Standard_D2s_v5", "Standard_D4s_v5", "Standard_D8s_v5", "Standard_D16s_v5", "Standard_D32s_v5", "Standard_D48s_v5", "Standard_D64s_v5", "Standard_D96s_v5", "Standard_D2_v5", "Standard_D4_v5", "Standard_D8_v5", "Standard_D16_v5", "Standard_D32_v5", "Standard_D48_v5", "Standard_D64_v5", "Standard_D96_v5", "Standard_E2ds_v5", "Standard_E4-2ds_v5", "Standard_E4ds_v5", "Standard_E8-2ds_v5", "Standard_E8-4ds_v5", "Standard_E8ds_v5", "Standard_E16-4ds_v5", "Standard_E16-8ds_v5", "Standard_E16ds_v5", "Standard_E20ds_v5", "Standard_E32-8ds_v5", "Standard_E32-16ds_v5", "Standard_E32ds_v5", "Standard_E48ds_v5", "Standard_E64-16ds_v5", "Standard_E64-32ds_v5", "Standard_E64ds_v5", "Standard_E96-24ds_v5", "Standard_E96-48ds_v5", "Standard_E96ds_v5", "Standard_E104ids_v5", "Standard_E2d_v5", "Standard_E4d_v5", "Standard_E8d_v5", "Standard_E16d_v5", "Standard_E20d_v5", "Standard_E32d_v5", "Standard_E48d_v5", "Standard_E64d_v5", "Standard_E96d_v5", "Standard_E104id_v5", "Standard_E2s_v5", "Standard_E4-2s_v5", "Standard_E4s_v5", "Standard_E8-2s_v5", "Standard_E8-4s_v5", "Standard_E8s_v5", "Standard_E16-4s_v5", "Standard_E16-8s_v5", "Standard_E16s_v5", "Standard_E20s_v5", "Standard_E32-8s_v5", "Standard_E32-16s_v5", "Standard_E32s_v5", "Standard_E48s_v5", "Standard_E64-16s_v5", "Standard_E64-32s_v5", "Standard_E64s_v5", "Standard_E96-24s_v5", "Standard_E96-48s_v5", "Standard_E96s_v5", "Standard_E104is_v5", "Standard_E2_v5", "Standard_E4_v5", "Standard_E8_v5", "Standard_E16_v5", "Standard_E20_v5", "Standard_E32_v5", "Standard_E48_v5", "Standard_E64_v5", "Standard_E96_v5", "Standard_E104i_v5", "Standard_E2bs_v5", "Standard_E4bs_v5", "Standard_E8bs_v5", "Standard_E16bs_v5", "Standard_E32bs_v5", "Standard_E48bs_v5", "Standard_E64bs_v5", "Standard_E2bds_v5", "Standard_E4bds_v5", "Standard_E8bds_v5", "Standard_E16bds_v5", "Standard_E32bds_v5", "Standard_E48bds_v5", "Standard_E64bds_v5", "Standard_D2ls_v5", "Standard_D4ls_v5", "Standard_D8ls_v5", "Standard_D16ls_v5", "Standard_D32ls_v5", "Standard_D48ls_v5", "Standard_D64ls_v5", "Standard_D96ls_v5", "Standard_D2lds_v5", "Standard_D4lds_v5", "Standard_D8lds_v5", "Standard_D16lds_v5", "Standard_D32lds_v5", "Standard_D48lds_v5", "Standard_D64lds_v5", "Standard_D96lds_v5", "Standard_L8s_v2", "Standard_L16s_v2", "Standard_L32s_v2", "Standard_L48s_v2", "Standard_L64s_v2", "Standard_L80s_v2", "Standard_NV4as_v4", "Standard_NV8as_v4", "Standard_NV16as_v4", "Standard_NV32as_v4", "Standard_L8s_v3", "Standard_L16s_v3", "Standard_L32s_v3", "Standard_L48s_v3", "Standard_L64s_v3", "Standard_L80s_v3", "Standard_E64i_v3", "Standard_E64is_v3", "Standard_G1", "Standard_G2", "Standard_G3", "Standard_G4", "Standard_G5", "Standard_GS1", "Standard_GS2", "Standard_GS3", "Standard_GS4", "Standard_GS4-4", "Standard_GS4-8", "Standard_GS5", "Standard_GS5-8", "Standard_GS5-16", "Standard_L4s", "Standard_L8s", "Standard_L16s", "Standard_L32s", "Standard_DC2as_v5", "Standard_DC4as_v5", "Standard_DC8as_v5", "Standard_DC16as_v5", "Standard_DC32as_v5", "Standard_DC48as_v5", "Standard_DC64as_v5", "Standard_DC96as_v5", "Standard_DC2ads_v5", "Standard_DC4ads_v5", "Standard_DC8ads_v5", "Standard_DC16ads_v5", "Standard_DC32ads_v5", "Standard_DC48ads_v5", "Standard_DC64ads_v5", "Standard_DC96ads_v5", "Standard_EC2as_v5", "Standard_EC4as_v5", "Standard_EC8as_v5", "Standard_EC16as_v5", "Standard_EC20as_v5", "Standard_EC32as_v5", "Standard_EC48as_v5", "Standard_EC64as_v5", "Standard_EC96as_v5", "Standard_EC96ias_v5", "Standard_EC2ads_v5", "Standard_EC4ads_v5", "Standard_EC8ads_v5", "Standard_EC16ads_v5", "Standard_EC20ads_v5", "Standard_EC32ads_v5", "Standard_EC48ads_v5", "Standard_EC64ads_v5", "Standard_EC96ads_v5", "Standard_EC96iads_v5", "Standard_E96ias_v4", "Standard_NC6s_v3", "Standard_NC12s_v3", "Standard_NC24rs_v3", "Standard_NC24s_v3", "Standard_NV6s_v2", "Standard_NV12s_v2", "Standard_NV24s_v2", "Standard_NV12s_v3", "Standard_NV24s_v3", "Standard_NV48s_v3", "Standard_M208ms_v2", "Standard_M208s_v2", "Standard_M416-208s_v2", "Standard_M416s_v2", "Standard_M416-208ms_v2", "Standard_M416ms_v2", "Standard_D2plds_v5", "Standard_D4plds_v5", "Standard_D8plds_v5", "Standard_D16plds_v5", "Standard_D32plds_v5", "Standard_D48plds_v5", "Standard_D64plds_v5", "Standard_D2pls_v5", "Standard_D4pls_v5", "Standard_D8pls_v5", "Standard_D16pls_v5", "Standard_D32pls_v5", "Standard_D48pls_v5", "Standard_D64pls_v5", "Standard_D2pds_v5", "Standard_D4pds_v5", "Standard_D8pds_v5", "Standard_D16pds_v5", "Standard_D32pds_v5", "Standard_D48pds_v5", "Standard_D64pds_v5", "Standard_D2ps_v5", "Standard_D4ps_v5", "Standard_D8ps_v5", "Standard_D16ps_v5", "Standard_D32ps_v5", "Standard_D48ps_v5", "Standard_D64ps_v5", "Standard_E2pds_v5", "Standard_E4pds_v5", "Standard_E8pds_v5", "Standard_E16pds_v5", "Standard_E20pds_v5", "Standard_E32pds_v5", "Standard_E2ps_v5", "Standard_E4ps_v5", "Standard_E8ps_v5", "Standard_E16ps_v5", "Standard_E20ps_v5", "Standard_E32ps_v5", "Standard_DC1s_v3", "Standard_DC2s_v3", "Standard_DC4s_v3", "Standard_DC8s_v3", "Standard_DC16s_v3", "Standard_DC24s_v3", "Standard_DC32s_v3", "Standard_DC48s_v3", "Standard_DC1ds_v3", "Standard_DC2ds_v3", "Standard_DC4ds_v3", "Standard_DC8ds_v3", "Standard_DC16ds_v3", "Standard_DC24ds_v3", "Standard_DC32ds_v3", "Standard_DC48ds_v3", "Standard_NC24ads_A100_v4", "Standard_NC48ads_A100_v4", "Standard_NC96ads_A100_v4",
	}
	for _, validRegion := range validDisks {
		if validRegion == size {
			return nil
		}
	}
	return fmt.Errorf("INVALID VM SIZE\nValid options: %v\n", validDisks)
}
