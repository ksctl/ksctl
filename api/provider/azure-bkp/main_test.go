package azure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	logger "github.com/kubesimplify/ksctl/api/provider/logger"
	util "github.com/kubesimplify/ksctl/api/provider/utils"
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

	logg := logger.Logger{Verbose: true}
	apiStore := util.AzureCredential{
		SubscriptionID: "demo-sub-id",
		TenantID:       "tenant-id",
		ClientID:       "client-id",
		ClientSecret:   "client-secret",
	}

	err := util.SaveCred(logg, apiStore, "azure")

	if err != nil {
		t.Fatalf("unable to save the credentials to file %v", err.Error())
	}

	setRequiredENV_VAR(logg, ctx, obj)
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		t.Fatalf("Failed to check %v", err.Error())
	}

	obj.AzureTokenCred = cred
	obj.Config = &AzureStateCluster{}
	obj.Config.ClusterName = obj.ClusterName
	obj.SSH_Payload = &util.SSHPayload{}
	err = setRequiredENV_VAR(logg, ctx, obj)
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

func TestAzureProvider_AddMoreWorkerNodes(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.AddMoreWorkerNodes(tt.args.logging); (err != nil) != tt.wantErr {
				t.Errorf("AddMoreWorkerNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_ConfigLoadBalancer(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		logger logger.Logger
		CPIPs  []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.ConfigLoadBalancer(tt.args.logger, tt.args.CPIPs); (err != nil) != tt.wantErr {
				t.Errorf("ConfigLoadBalancer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_ConfigReader(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		logging     logger.Logger
		clusterType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := config.ConfigReader(tt.args.logging, tt.args.clusterType); (err != nil) != tt.wantErr {
				t.Errorf("ConfigReader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_ConfigWriter(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		logging     logger.Logger
		clusterType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := config.ConfigWriter(tt.args.logging, tt.args.clusterType); (err != nil) != tt.wantErr {
				t.Errorf("ConfigWriter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_CreateCluster(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.CreateCluster(tt.args.logging); (err != nil) != tt.wantErr {
				t.Errorf("CreateCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_CreateNSG(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx           context.Context
		logging       logger.Logger
		nsgName       string
		securityRules []*armnetwork.SecurityRule
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *armnetwork.SecurityGroup
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			got, err := obj.CreateNSG(tt.args.ctx, tt.args.logging, tt.args.nsgName, tt.args.securityRules)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateNSG() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateNSG() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureProvider_CreateNetworkInterface(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx                    context.Context
		logging                logger.Logger
		resourceName           string
		nicName                string
		subnetID               string
		publicIPID             string
		networkSecurityGroupID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *armnetwork.Interface
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			got, err := obj.CreateNetworkInterface(tt.args.ctx, tt.args.logging, tt.args.resourceName, tt.args.nicName, tt.args.subnetID, tt.args.publicIPID, tt.args.networkSecurityGroupID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateNetworkInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateNetworkInterface() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureProvider_CreatePublicIP(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx          context.Context
		logging      logger.Logger
		publicIPName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *armnetwork.PublicIPAddress
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			got, err := obj.CreatePublicIP(tt.args.ctx, tt.args.logging, tt.args.publicIPName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePublicIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreatePublicIP() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureProvider_CreateResourceGroup(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx     context.Context
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *armresources.ResourceGroupsClientCreateOrUpdateResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			got, err := obj.CreateResourceGroup(tt.args.ctx, tt.args.logging)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateResourceGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateResourceGroup() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureProvider_CreateSubnet(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx        context.Context
		logging    logger.Logger
		subnetName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *armnetwork.Subnet
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			got, err := obj.CreateSubnet(tt.args.ctx, tt.args.logging, tt.args.subnetName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSubnet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateSubnet() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureProvider_CreateVM(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx                context.Context
		logging            logger.Logger
		vmName             string
		networkInterfaceID string
		diskName           string
		script             string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *armcompute.VirtualMachine
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			got, err := obj.CreateVM(tt.args.ctx, tt.args.logging, tt.args.vmName, tt.args.networkInterfaceID, tt.args.diskName, tt.args.script)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateVM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateVM() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureProvider_CreateVirtualNetwork(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx                context.Context
		logging            logger.Logger
		virtualNetworkName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *armnetwork.VirtualNetwork
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			got, err := obj.CreateVirtualNetwork(tt.args.ctx, tt.args.logging, tt.args.virtualNetworkName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateVirtualNetwork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateVirtualNetwork() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureProvider_DeleteAllDisks(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx     context.Context
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeleteAllDisks(tt.args.ctx, tt.args.logging); (err != nil) != tt.wantErr {
				t.Errorf("DeleteAllDisks() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_DeleteAllNSG(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx     context.Context
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeleteAllNSG(tt.args.ctx, tt.args.logging); (err != nil) != tt.wantErr {
				t.Errorf("DeleteAllNSG() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_DeleteAllNetworkInterface(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx     context.Context
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeleteAllNetworkInterface(tt.args.ctx, tt.args.logging); (err != nil) != tt.wantErr {
				t.Errorf("DeleteAllNetworkInterface() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_DeleteAllPublicIP(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx     context.Context
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeleteAllPublicIP(tt.args.ctx, tt.args.logging); (err != nil) != tt.wantErr {
				t.Errorf("DeleteAllPublicIP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_DeleteAllVMs(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx     context.Context
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeleteAllVMs(tt.args.ctx, tt.args.logging); (err != nil) != tt.wantErr {
				t.Errorf("DeleteAllVMs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_DeleteCluster(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeleteCluster(tt.args.logging); (err != nil) != tt.wantErr {
				t.Errorf("DeleteCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_DeleteDisk(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx      context.Context
		logging  logger.Logger
		diskName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeleteDisk(tt.args.ctx, tt.args.logging, tt.args.diskName); (err != nil) != tt.wantErr {
				t.Errorf("DeleteDisk() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_DeleteNSG(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx     context.Context
		logging logger.Logger
		nsgName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeleteNSG(tt.args.ctx, tt.args.logging, tt.args.nsgName); (err != nil) != tt.wantErr {
				t.Errorf("DeleteNSG() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_DeleteNetworkInterface(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx     context.Context
		logging logger.Logger
		nicName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeleteNetworkInterface(tt.args.ctx, tt.args.logging, tt.args.nicName); (err != nil) != tt.wantErr {
				t.Errorf("DeleteNetworkInterface() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_DeletePublicIP(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx          context.Context
		logging      logger.Logger
		publicIPName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeletePublicIP(tt.args.ctx, tt.args.logging, tt.args.publicIPName); (err != nil) != tt.wantErr {
				t.Errorf("DeletePublicIP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_DeleteResourceGroup(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx     context.Context
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeleteResourceGroup(tt.args.ctx, tt.args.logging); (err != nil) != tt.wantErr {
				t.Errorf("DeleteResourceGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_DeleteSSHKeyPair(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx     context.Context
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeleteSSHKeyPair(tt.args.ctx, tt.args.logging); (err != nil) != tt.wantErr {
				t.Errorf("DeleteSSHKeyPair() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_DeleteSomeWorkerNodes(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeleteSomeWorkerNodes(tt.args.logging); (err != nil) != tt.wantErr {
				t.Errorf("DeleteSomeWorkerNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_DeleteSubnet(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx        context.Context
		logging    logger.Logger
		subnetName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeleteSubnet(tt.args.ctx, tt.args.logging, tt.args.subnetName); (err != nil) != tt.wantErr {
				t.Errorf("DeleteSubnet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_DeleteVM(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx     context.Context
		logging logger.Logger
		vmName  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeleteVM(tt.args.ctx, tt.args.logging, tt.args.vmName); (err != nil) != tt.wantErr {
				t.Errorf("DeleteVM() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_DeleteVirtualNetwork(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx     context.Context
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.DeleteVirtualNetwork(tt.args.ctx, tt.args.logging); (err != nil) != tt.wantErr {
				t.Errorf("DeleteVirtualNetwork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_FetchKUBECONFIG(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		logging  logger.Logger
		publicIP string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			got, err := obj.FetchKUBECONFIG(tt.args.logging, tt.args.publicIP)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchKUBECONFIG() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FetchKUBECONFIG() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureProvider_GetTokenFromCP_1(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		logger   logger.Logger
		PublicIP string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if got := obj.GetTokenFromCP_1(tt.args.logger, tt.args.PublicIP); got != tt.want {
				t.Errorf("GetTokenFromCP_1() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureProvider_HelperExecNoOutputControlPlane(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		logger   logger.Logger
		publicIP string
		script   string
		fastMode bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.HelperExecNoOutputControlPlane(tt.args.logger, tt.args.publicIP, tt.args.script, tt.args.fastMode); (err != nil) != tt.wantErr {
				t.Errorf("HelperExecNoOutputControlPlane() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_HelperExecOutputControlPlane(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		logger   logger.Logger
		publicIP string
		script   string
		fastMode bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			got, err := obj.HelperExecOutputControlPlane(tt.args.logger, tt.args.publicIP, tt.args.script, tt.args.fastMode)
			if (err != nil) != tt.wantErr {
				t.Errorf("HelperExecOutputControlPlane() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HelperExecOutputControlPlane() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureProvider_SaveKubeconfig(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		logging    logger.Logger
		kubeconfig string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.SaveKubeconfig(tt.args.logging, tt.args.kubeconfig); (err != nil) != tt.wantErr {
				t.Errorf("SaveKubeconfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_SwitchContext(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	log := logger.Logger{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := provider.SwitchContext(log); (err != nil) != tt.wantErr {
				t.Errorf("SwitchContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_UploadSSHKey(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.UploadSSHKey(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("UploadSSHKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_createControlPlane(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx         context.Context
		logger      logger.Logger
		indexOfNode int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.createControlPlane(tt.args.ctx, tt.args.logger, tt.args.indexOfNode); (err != nil) != tt.wantErr {
				t.Errorf("createControlPlane() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_createDatabase(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		ctx    context.Context
		logger logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.createDatabase(tt.args.ctx, tt.args.logger); (err != nil) != tt.wantErr {
				t.Errorf("createDatabase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_createLoadBalancer(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		logger logger.Logger
		ctx    context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.createLoadBalancer(tt.args.logger, tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("createLoadBalancer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_createWorkerPlane(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	type args struct {
		logger      logger.Logger
		ctx         context.Context
		indexOfNode int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			if err := obj.createWorkerPlane(tt.args.logger, tt.args.ctx, tt.args.indexOfNode); (err != nil) != tt.wantErr {
				t.Errorf("createWorkerPlane() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureProvider_kubeconfigReader(t *testing.T) {
	type fields struct {
		ClusterName    string
		HACluster      bool
		Region         string
		Spec           util.Machine
		SubscriptionID string
		Config         *AzureStateCluster
		AzureTokenCred azcore.TokenCredential
		SSH_Payload    *util.SSHPayload
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &AzureProvider{
				ClusterName:    tt.fields.ClusterName,
				HACluster:      tt.fields.HACluster,
				Region:         tt.fields.Region,
				Spec:           tt.fields.Spec,
				SubscriptionID: tt.fields.SubscriptionID,
				Config:         tt.fields.Config,
				AzureTokenCred: tt.fields.AzureTokenCred,
				SSH_Payload:    tt.fields.SSH_Payload,
			}
			got, err := obj.kubeconfigReader()
			if (err != nil) != tt.wantErr {
				t.Errorf("kubeconfigReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("kubeconfigReader() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentials(t *testing.T) {
	type args struct {
		logger logger.Logger
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Credentials(tt.args.logger); got != tt.want {
				t.Errorf("Credentials() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_configLBscript(t *testing.T) {
	type args struct {
		controlPlaneIPs []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := configLBscript(tt.args.controlPlaneIPs); got != tt.want {
				t.Errorf("configLBscript() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateDBPassword(t *testing.T) {
	type args struct {
		passwordLen int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateDBPassword(tt.args.passwordLen); got != tt.want {
				t.Errorf("generateDBPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getAzureManagedClusterClient(t *testing.T) {
	type args struct {
		cred *AzureProvider
	}
	tests := []struct {
		name    string
		args    args
		want    *armcontainerservice.ManagedClustersClient
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getAzureManagedClusterClient(tt.args.cred)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAzureManagedClusterClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAzureManagedClusterClient() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getAzureResourceGroupsClient(t *testing.T) {
	type args struct {
		cred *AzureProvider
	}
	tests := []struct {
		name    string
		args    args
		want    *armresources.ResourceGroupsClient
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getAzureResourceGroupsClient(tt.args.cred)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAzureResourceGroupsClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAzureResourceGroupsClient() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getControlPlaneFirewallRules(t *testing.T) {
	tests := []struct {
		name              string
		wantSecurityRules []*armnetwork.SecurityRule
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSecurityRules := getControlPlaneFirewallRules(); !reflect.DeepEqual(gotSecurityRules, tt.wantSecurityRules) {
				t.Errorf("getControlPlaneFirewallRules() = %v, want %v", gotSecurityRules, tt.wantSecurityRules)
			}
		})
	}
}

func Test_getDatabaseFirewallRules(t *testing.T) {
	tests := []struct {
		name              string
		wantSecurityRules []*armnetwork.SecurityRule
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSecurityRules := getDatabaseFirewallRules(); !reflect.DeepEqual(gotSecurityRules, tt.wantSecurityRules) {
				t.Errorf("getDatabaseFirewallRules() = %v, want %v", gotSecurityRules, tt.wantSecurityRules)
			}
		})
	}
}

func Test_getLoadBalancerFirewallRules(t *testing.T) {
	tests := []struct {
		name              string
		wantSecurityRules []*armnetwork.SecurityRule
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSecurityRules := getLoadBalancerFirewallRules(); !reflect.DeepEqual(gotSecurityRules, tt.wantSecurityRules) {
				t.Errorf("getLoadBalancerFirewallRules() = %v, want %v", gotSecurityRules, tt.wantSecurityRules)
			}
		})
	}
}

func Test_getWorkerPlaneFirewallRules(t *testing.T) {
	tests := []struct {
		name              string
		wantSecurityRules []*armnetwork.SecurityRule
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSecurityRules := getWorkerPlaneFirewallRules(); !reflect.DeepEqual(gotSecurityRules, tt.wantSecurityRules) {
				t.Errorf("getWorkerPlaneFirewallRules() = %v, want %v", gotSecurityRules, tt.wantSecurityRules)
			}
		})
	}
}

func Test_haCreateClusterHandler(t *testing.T) {
	type args struct {
		ctx    context.Context
		logger logger.Logger
		obj    *AzureProvider
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := haCreateClusterHandler(tt.args.ctx, tt.args.logger, tt.args.obj); (err != nil) != tt.wantErr {
				t.Errorf("haCreateClusterHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_haDeleteClusterHandler(t *testing.T) {
	type args struct {
		ctx     context.Context
		logger  logger.Logger
		obj     *AzureProvider
		showMsg bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := haDeleteClusterHandler(tt.args.ctx, tt.args.logger, tt.args.obj, tt.args.showMsg); (err != nil) != tt.wantErr {
				t.Errorf("haDeleteClusterHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_isPresent(t *testing.T) {
	type args struct {
		kind string
		obj  AzureProvider
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPresent(tt.args.kind, tt.args.obj); got != tt.want {
				t.Errorf("isPresent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_managedCreateClusterHandler(t *testing.T) {
	type args struct {
		ctx         context.Context
		logging     logger.Logger
		azureConfig *AzureProvider
	}
	tests := []struct {
		name    string
		args    args
		want    *armcontainerservice.ManagedCluster
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := managedCreateClusterHandler(tt.args.ctx, tt.args.logging, tt.args.azureConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("managedCreateClusterHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("managedCreateClusterHandler() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_managedDeleteClusterHandler(t *testing.T) {
	type args struct {
		ctx         context.Context
		logging     logger.Logger
		azureConfig *AzureProvider
		showMsg     bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := managedDeleteClusterHandler(tt.args.ctx, tt.args.logging, tt.args.azureConfig, tt.args.showMsg); (err != nil) != tt.wantErr {
				t.Errorf("managedDeleteClusterHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_printer_Printer(t *testing.T) {
	type fields struct {
		ClusterName  string
		Region       string
		ResourceName string
	}
	type args struct {
		isHA      bool
		operation int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	log := logger.Logger{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := printer{
				ClusterName:  tt.fields.ClusterName,
				Region:       tt.fields.Region,
				ResourceName: tt.fields.ResourceName,
			}
			p.Printer(log, tt.args.isHA, tt.args.operation)
		})
	}
}

func Test_scriptCP_n(t *testing.T) {
	type args struct {
		dbEndpoint  string
		privateIPlb string
		token       string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scriptCP_n(tt.args.dbEndpoint, tt.args.privateIPlb, tt.args.token); got != tt.want {
				t.Errorf("scriptCP_n() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scriptDB(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scriptDB(tt.args.password); got != tt.want {
				t.Errorf("scriptDB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scriptKUBECONFIG(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scriptKUBECONFIG(); got != tt.want {
				t.Errorf("scriptKUBECONFIG() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scriptLB(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scriptLB(); got != tt.want {
				t.Errorf("scriptLB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scriptWP(t *testing.T) {
	type args struct {
		privateIPlb string
		token       string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scriptWP(tt.args.privateIPlb, tt.args.token); got != tt.want {
				t.Errorf("scriptWP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scriptWithCP_1(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scriptWithCP_1(); got != tt.want {
				t.Errorf("scriptWithCP_1() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scriptWithoutCP_1(t *testing.T) {
	type args struct {
		dbEndpoint  string
		privateIPlb string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scriptWithoutCP_1(tt.args.dbEndpoint, tt.args.privateIPlb); got != tt.want {
				t.Errorf("scriptWithoutCP_1() = %v, want %v", got, tt.want)
			}
		})
	}
}
