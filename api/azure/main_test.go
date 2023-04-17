package azure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	log "github.com/kubesimplify/ksctl/api/logger"
	util "github.com/kubesimplify/ksctl/api/utils"
)

func TestValidRegions(t *testing.T) {
	testData := []string{"abcd", "eastus", "westus2"}
	expectedResult := []bool{false, true, true}
	for i := 0; i < len(testData); i++ {
		if isValidRegion(testData[i]) != expectedResult[i] {
			t.Fatalf("%s region got %v but was expecting %v", testData[i], isValidRegion(testData[i]), expectedResult[i])
		}
	}
}

func TestValidNodeSizes(t *testing.T) {
	testData := []string{"Standard_D16d_v5", "Standard_F1", "Standard_sdcdd"}
	expectedResult := []bool{true, true, false}
	for i := 0; i < len(testData); i++ {
		if isValidNodeSize(testData[i]) != expectedResult[i] {
			t.Fatalf("%s region got %v but was expecting %v", testData[i], isValidNodeSize(testData[i]), expectedResult[i])
		}
	}
}
func setup() {
	err := os.MkdirAll(util.GetPath(util.CLUSTER_PATH, "civo", "managed"), 0750)
	if err != nil {
		return
	}
	err = os.MkdirAll(fmt.Sprintf("%s/.ksctl/cred", util.GetUserName()), 0755)

	if err != nil {
		return
	}
}

func cleanup() {
	fmt.Println("Cleanup triggered")
	_ = os.RemoveAll(fmt.Sprintf("%s/.ksctl", util.GetUserName()))
}

// TODO: update this test case once the file and environemnt variable
// both are used for getting the secret keys
func TestSettingEnvironmentVariables(t *testing.T) {
	ctx := context.Background()
	obj := &AzureProvider{
		ClusterName: "demo",
		Region:      "",
	}
	setup()
	defer cleanup()

	logger := log.Logger{Verbose: true}
	apiStore := util.AzureCredential{
		SubscriptionID: "demo-sub-id",
		TenantID:       "tenant-id",
		ClientID:       "client-id",
		ClientSecret:   "client-secret",
	}

	err := util.SaveCred(logger, apiStore, "azure")

	if err != nil {
		t.Fatalf("unable to save the credentials to file %v", err.Error())
	}

	setRequiredENV_VAR(logger, ctx, obj)
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		t.Fatalf("Failed to check %v", err.Error())
	}

	obj.AzureTokenCred = cred
	obj.Config = &AzureStateCluster{}
	obj.Config.ClusterName = obj.ClusterName
	obj.SSH_Payload = &util.SSHPayload{}
	err = setRequiredENV_VAR(logger, ctx, obj)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Fatalf("File not found : %v", err.Error())
		}
		t.Fatalf("Failed to set the environment variables %v", err.Error())
	}

	var (
		subscription  bool
		client_id     bool
		client_secret bool
		tenant_id     bool
	)
	if len(obj.SubscriptionID) != 0 && obj.SubscriptionID == apiStore.SubscriptionID {
		subscription = true
	}
	if tenant := os.Getenv("AZURE_TENANT_ID"); len(tenant) != 0 && tenant == apiStore.TenantID {
		tenant_id = true
	}
	if client_i := os.Getenv("AZURE_CLIENT_ID"); len(client_i) != 0 && client_i == apiStore.ClientID {
		client_id = true
	}
	if client_s := os.Getenv("AZURE_CLIENT_SECRET"); len(client_s) != 0 && client_s == apiStore.ClientSecret {
		client_secret = true
	}

	if !subscription || !client_id || !client_secret || !tenant_id {
		t.Fatalf("Some environment variables where not set")
	}
}
