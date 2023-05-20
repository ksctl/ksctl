package civo

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/civo/civogo"

	"github.com/kubesimplify/ksctl/api/logger"
	utils "github.com/kubesimplify/ksctl/api/utils"
	"github.com/stretchr/testify/assert"
)

func TestFetchAPIKey(T *testing.T) {
	logging := logger.Logger{}

	apikey := fetchAPIKey(logging)

	if fmt.Sprintf("%T", apikey) != "string" || len(apikey) != 0 {
		T.Fatalf("Invalid Return type or APIKey already present")
	}
}

func TestIsValidNodeSize(T *testing.T) {
	validSizes := []string{"g4s.kube.xsmall", "g4s.kube.small", "g4s.kube.medium", "g4s.kube.large", "g4p.kube.small", "g4p.kube.medium", "g4p.kube.large", "g4p.kube.xlarge", "g4c.kube.small", "g4c.kube.medium", "g4c.kube.large", "g4c.kube.xlarge", "g4m.kube.small", "g4m.kube.medium", "g4m.kube.large", "g4m.kube.xlarge"}
	testData := validSizes[rand.Int()%len(validSizes)]
	assert.Equalf(T, true, isValidSizeManaged(testData), "Returns False for valid size")

	assert.Equalf(T, false, isValidSizeManaged("abcd"), "Returns True for invalid node size")
	assert.Equalf(T, false, isValidSizeManaged("kube.small"), "Returns True for invalid node size")
	assert.Equalf(T, false, isValidSizeManaged("g4s.k3s.small"), "Returns True for invalid node size")
}

//TODO: Test ClusterInfoInjecter()

//TODO: Test kubeconfigDeleter()

//Testing of deleteClusterWithID() and DeleteCluster() and CreateCluster() [TODO Need to be done]

func setup() {
	err := os.MkdirAll(utils.GetPath(utils.CLUSTER_PATH, "civo", "managed"), 0750)
	if err != nil {
		return
	}
}

func TestIsPresent(t *testing.T) {
	setup()
	t.Cleanup(func() {
		_ = os.RemoveAll(utils.GetPath(utils.CLUSTER_PATH, "civo"))
	})

	present := isPresent("managed", "demo", "LON1")
	assert.Equal(t, false, present, "with no clusters returns true! (false +ve)")
	err := os.Mkdir(utils.GetPath(utils.CLUSTER_PATH, "civo", "managed", "demo LON1"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Create(utils.GetPath(utils.CLUSTER_PATH, "civo", "managed", "demo LON1", "info.json"))
	if err != nil {
		t.Fatal(err)
	}
	present = isPresent("managed", "demo", "LON1")
	assert.Equal(t, true, present, "Failed to detect the cluster (false -ve)")
}

func TestClusterInfoInjecterManagedType(t *testing.T) {
	clusterName := "xYz"
	region := "aBc"
	nodeSize := "k3s"
	logging := logger.Logger{}
	abcd := ClusterInfoInjecter(logging, clusterName, region, nodeSize, 1, "", "")
	worker := CivoProvider{
		ClusterName: clusterName,
		Region:      region,
		Spec:        utils.Machine{Disk: nodeSize, ManagedNodes: 1},
		APIKey:      fetchAPIKey(logging),
		Application: "Traefik-v2-nodeport,metrics-server", // EXPLICITLY mentioned the expected data
		CNIPlugin:   "flannel",
	}
	if worker != abcd {
		t.Fatalf("Base check failed")
	}
}

// testing civo components

var (
	civoOperator HACollection
)

func InitTesting_HA(t *testing.T) {
	civoOperator = &HAType{
		Client:        nil,
		NodeSize:      "",
		ClusterName:   "demo",
		DiskImgID:     "id",
		DBFirewallID:  "",
		LBFirewallID:  "",
		CPFirewallID:  "",
		WPFirewallID:  "",
		NetworkID:     "",
		SSHID:         "",
		Configuration: nil,
		SSH_Payload:   nil,
	}
}

func TestSwitchContext(t *testing.T) {

	t.Cleanup(func() {
		_ = os.RemoveAll(utils.GetPath(utils.OTHER_PATH, "civo"))
	})

	if err := os.MkdirAll(utils.GetPath(utils.CLUSTER_PATH, "civo", "managed", "demo-1 FRA1"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(utils.GetPath(utils.CLUSTER_PATH, "civo", "ha", "demo-2 LON1"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(utils.GetPath(utils.CLUSTER_PATH, "civo", "ha", "demo-2 LON1", "info.json"), []byte("{}"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(utils.GetPath(utils.CLUSTER_PATH, "civo", "managed", "demo-1 FRA1", "info.json"), []byte("{}"), 0755); err != nil {
		t.Fatal(err)
	}

	civoOperator := CivoProvider{
		ClusterName: "demo",
		Region:      "Abcd",
	}
	log := logger.Logger{}
	if err := civoOperator.SwitchContext(log); err == nil {
		t.Fatalf("Passed when their is no matching cluster")
	}
	civoOperator.ClusterName = "demo-1"
	civoOperator.Region = "FRA1"
	civoOperator.HACluster = false

	if err := civoOperator.SwitchContext(log); err != nil {
		t.Fatalf("Failed in switching context to %v\nError: %v\n", civoOperator, err)
	}

	civoOperator.ClusterName = "demo-2"
	civoOperator.Region = "LON1"
	civoOperator.HACluster = true

	if err := civoOperator.SwitchContext(log); err != nil {
		t.Fatalf("Failed in switching context to %v\nError: %v\n", civoOperator, err)
	}
}

func TestUploadSSHKey(t *testing.T) {}

func TestCreateDatabase(t *testing.T) {}

func TestCreateLoadbalancer(t *testing.T) {}

func TestCreateControlPlane(t *testing.T) {}

func TestCreateWorkerNode(t *testing.T) {}

//import (
//"fmt"
//"github.com/civo/civogo"
//"github.com/kubesimplify/ksctl/api/logger"
//"github.com/kubesimplify/ksctl/api/utils"
//"github.com/stretchr/testify/assert"
//"testing"
//)

func TestCivoProvider_AddMoreWorkerNodes(t *testing.T) {
	type fields struct {
		ClusterName string
		APIKey      string
		HACluster   bool
		Region      string
		Spec        utils.Machine
		Application string
		CNIPlugin   string
	}
	type args struct {
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := CivoProvider{
				ClusterName: tt.fields.ClusterName,
				APIKey:      tt.fields.APIKey,
				HACluster:   tt.fields.HACluster,
				Region:      tt.fields.Region,
				Spec:        tt.fields.Spec,
				Application: tt.fields.Application,
				CNIPlugin:   tt.fields.CNIPlugin,
			}
			tt.wantErr(t, provider.AddMoreWorkerNodes(tt.args.logging), fmt.Sprintf("AddMoreWorkerNodes(%v)", tt.args.logging))
		})
	}
}

func TestCivoProvider_CreateCluster(t *testing.T) {
	type fields struct {
		ClusterName string
		APIKey      string
		HACluster   bool
		Region      string
		Spec        utils.Machine
		Application string
		CNIPlugin   string
	}
	type args struct {
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := CivoProvider{
				ClusterName: tt.fields.ClusterName,
				APIKey:      tt.fields.APIKey,
				HACluster:   tt.fields.HACluster,
				Region:      tt.fields.Region,
				Spec:        tt.fields.Spec,
				Application: tt.fields.Application,
				CNIPlugin:   tt.fields.CNIPlugin,
			}
			tt.wantErr(t, provider.CreateCluster(tt.args.logging), fmt.Sprintf("CreateCluster(%v)", tt.args.logging))
		})
	}
}

func TestCivoProvider_DeleteCluster(t *testing.T) {
	type fields struct {
		ClusterName string
		APIKey      string
		HACluster   bool
		Region      string
		Spec        utils.Machine
		Application string
		CNIPlugin   string
	}
	type args struct {
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := CivoProvider{
				ClusterName: tt.fields.ClusterName,
				APIKey:      tt.fields.APIKey,
				HACluster:   tt.fields.HACluster,
				Region:      tt.fields.Region,
				Spec:        tt.fields.Spec,
				Application: tt.fields.Application,
				CNIPlugin:   tt.fields.CNIPlugin,
			}
			tt.wantErr(t, provider.DeleteCluster(tt.args.logging), fmt.Sprintf("DeleteCluster(%v)", tt.args.logging))
		})
	}
}

func TestCivoProvider_DeleteSomeWorkerNodes(t *testing.T) {
	type fields struct {
		ClusterName string
		APIKey      string
		HACluster   bool
		Region      string
		Spec        utils.Machine
		Application string
		CNIPlugin   string
	}
	type args struct {
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := CivoProvider{
				ClusterName: tt.fields.ClusterName,
				APIKey:      tt.fields.APIKey,
				HACluster:   tt.fields.HACluster,
				Region:      tt.fields.Region,
				Spec:        tt.fields.Spec,
				Application: tt.fields.Application,
				CNIPlugin:   tt.fields.CNIPlugin,
			}
			tt.wantErr(t, provider.DeleteSomeWorkerNodes(tt.args.logging), fmt.Sprintf("DeleteSomeWorkerNodes(%v)", tt.args.logging))
		})
	}
}

func TestCivoProvider_SwitchContext(t *testing.T) {
	type fields struct {
		ClusterName string
		APIKey      string
		HACluster   bool
		Region      string
		Spec        utils.Machine
		Application string
		CNIPlugin   string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	log := logger.Logger{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := CivoProvider{
				ClusterName: tt.fields.ClusterName,
				APIKey:      tt.fields.APIKey,
				HACluster:   tt.fields.HACluster,
				Region:      tt.fields.Region,
				Spec:        tt.fields.Spec,
				Application: tt.fields.Application,
				CNIPlugin:   tt.fields.CNIPlugin,
			}
			tt.wantErr(t, provider.SwitchContext(log), fmt.Sprintf("SwitchContext()"))
		})
	}
}

func TestClusterInfoInjecter(t *testing.T) {
	type args struct {
		logging     logger.Logger
		clusterName string
		reg         string
		size        string
		noOfNodes   int
		application string
		cniPlugin   string
	}
	tests := []struct {
		name string
		args args
		want CivoProvider
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ClusterInfoInjecter(tt.args.logging, tt.args.clusterName, tt.args.reg, tt.args.size, tt.args.noOfNodes, tt.args.application, tt.args.cniPlugin), "ClusterInfoInjecter(%v, %v, %v, %v, %v, %v, %v)", tt.args.logging, tt.args.clusterName, tt.args.reg, tt.args.size, tt.args.noOfNodes, tt.args.application, tt.args.cniPlugin)
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
			assert.Equalf(t, tt.want, Credentials(tt.args.logger), "Credentials(%v)", tt.args.logger)
		})
	}
}

func TestDeleteAllPaths(t *testing.T) {
	type args struct {
		clusterName string
		region      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, DeleteAllPaths(tt.args.clusterName, tt.args.region), fmt.Sprintf("DeleteAllPaths(%v, %v)", tt.args.clusterName, tt.args.region))
		})
	}
}

func TestExtractInstances(t *testing.T) {
	type args struct {
		clusterName string
		region      string
	}
	tests := []struct {
		name        string
		args        args
		wantInstIDs InstanceID
		wantErr     assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInstIDs, err := ExtractInstances(tt.args.clusterName, tt.args.region)
			if !tt.wantErr(t, err, fmt.Sprintf("ExtractInstances(%v, %v)", tt.args.clusterName, tt.args.region)) {
				return
			}
			assert.Equalf(t, tt.wantInstIDs, gotInstIDs, "ExtractInstances(%v, %v)", tt.args.clusterName, tt.args.region)
		})
	}
}

func TestExtractNetworks(t *testing.T) {
	type args struct {
		clusterName string
		region      string
	}
	tests := []struct {
		name        string
		args        args
		wantInstIDs NetworkID
		wantErr     assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInstIDs, err := ExtractNetworks(tt.args.clusterName, tt.args.region)
			if !tt.wantErr(t, err, fmt.Sprintf("ExtractNetworks(%v, %v)", tt.args.clusterName, tt.args.region)) {
				return
			}
			assert.Equalf(t, tt.wantInstIDs, gotInstIDs, "ExtractNetworks(%v, %v)", tt.args.clusterName, tt.args.region)
		})
	}
}

func TestGetConfig(t *testing.T) {
	type args struct {
		clusterName string
		region      string
	}
	tests := []struct {
		name            string
		args            args
		wantConfigStore JsonStore
		wantErr         assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfigStore, err := GetConfig(tt.args.clusterName, tt.args.region)
			if !tt.wantErr(t, err, fmt.Sprintf("GetConfig(%v, %v)", tt.args.clusterName, tt.args.region)) {
				return
			}
			assert.Equalf(t, tt.wantConfigStore, gotConfigStore, "GetConfig(%v, %v)", tt.args.clusterName, tt.args.region)
		})
	}
}

func TestGetConfigManaged(t *testing.T) {
	type args struct {
		clusterName string
		region      string
	}
	tests := []struct {
		name            string
		args            args
		wantConfigStore ManagedConfig
		wantErr         assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfigStore, err := GetConfigManaged(tt.args.clusterName, tt.args.region)
			if !tt.wantErr(t, err, fmt.Sprintf("GetConfigManaged(%v, %v)", tt.args.clusterName, tt.args.region)) {
				return
			}
			assert.Equalf(t, tt.wantConfigStore, gotConfigStore, "GetConfigManaged(%v, %v)", tt.args.clusterName, tt.args.region)
		})
	}
}

func TestHAType_ConfigLoadBalancer(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		logging  logger.Logger
		instance *civogo.Instance
		CPIPs    []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			tt.wantErr(t, obj.ConfigLoadBalancer(tt.args.logging, tt.args.instance, tt.args.CPIPs), fmt.Sprintf("ConfigLoadBalancer(%v, %v, %v)", tt.args.logging, tt.args.instance, tt.args.CPIPs))
		})
	}
}

func TestHAType_CreateControlPlane(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		logging logger.Logger
		number  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *civogo.Instance
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			got, err := obj.CreateControlPlane(tt.args.logging, tt.args.number)
			if !tt.wantErr(t, err, fmt.Sprintf("CreateControlPlane(%v, %v)", tt.args.logging, tt.args.number)) {
				return
			}
			assert.Equalf(t, tt.want, got, "CreateControlPlane(%v, %v)", tt.args.logging, tt.args.number)
		})
	}
}

func TestHAType_CreateDatabase(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			got, err := obj.CreateDatabase(tt.args.logging)
			if !tt.wantErr(t, err, fmt.Sprintf("CreateDatabase(%v)", tt.args.logging)) {
				return
			}
			assert.Equalf(t, tt.want, got, "CreateDatabase(%v)", tt.args.logging)
		})
	}
}

func TestHAType_CreateFirewall(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		firewallName string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantFirew *civogo.FirewallResult
		wantErr   assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			gotFirew, err := obj.CreateFirewall(tt.args.firewallName)
			if !tt.wantErr(t, err, fmt.Sprintf("CreateFirewall(%v)", tt.args.firewallName)) {
				return
			}
			assert.Equalf(t, tt.wantFirew, gotFirew, "CreateFirewall(%v)", tt.args.firewallName)
		})
	}
}

func TestHAType_CreateInstance(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		instanceName         string
		firewallID           string
		NodeSize             string
		initializationScript string
		public               bool
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantInst *civogo.Instance
		wantErr  assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			gotInst, err := obj.CreateInstance(tt.args.instanceName, tt.args.firewallID, tt.args.NodeSize, tt.args.initializationScript, tt.args.public)
			if !tt.wantErr(t, err, fmt.Sprintf("CreateInstance(%v, %v, %v, %v, %v)", tt.args.instanceName, tt.args.firewallID, tt.args.NodeSize, tt.args.initializationScript, tt.args.public)) {
				return
			}
			assert.Equalf(t, tt.wantInst, gotInst, "CreateInstance(%v, %v, %v, %v, %v)", tt.args.instanceName, tt.args.firewallID, tt.args.NodeSize, tt.args.initializationScript, tt.args.public)
		})
	}
}

func TestHAType_CreateLoadbalancer(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *civogo.Instance
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			got, err := obj.CreateLoadbalancer(tt.args.logging)
			if !tt.wantErr(t, err, fmt.Sprintf("CreateLoadbalancer(%v)", tt.args.logging)) {
				return
			}
			assert.Equalf(t, tt.want, got, "CreateLoadbalancer(%v)", tt.args.logging)
		})
	}
}

func TestHAType_CreateNetwork(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		logging     logger.Logger
		networkName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			tt.wantErr(t, obj.CreateNetwork(tt.args.logging, tt.args.networkName), fmt.Sprintf("CreateNetwork(%v, %v)", tt.args.logging, tt.args.networkName))
		})
	}
}

func TestHAType_CreateSSHKeyPair(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		logging   logger.Logger
		publicKey string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			tt.wantErr(t, obj.CreateSSHKeyPair(tt.args.logging, tt.args.publicKey), fmt.Sprintf("CreateSSHKeyPair(%v, %v)", tt.args.logging, tt.args.publicKey))
		})
	}
}

func TestHAType_CreateWorkerNode(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		logging     logger.Logger
		number      int
		privateIPlb string
		token       string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *civogo.Instance
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			got, err := obj.CreateWorkerNode(tt.args.logging, tt.args.number, tt.args.privateIPlb, tt.args.token)
			if !tt.wantErr(t, err, fmt.Sprintf("CreateWorkerNode(%v, %v, %v, %v)", tt.args.logging, tt.args.number, tt.args.privateIPlb, tt.args.token)) {
				return
			}
			assert.Equalf(t, tt.want, got, "CreateWorkerNode(%v, %v, %v, %v)", tt.args.logging, tt.args.number, tt.args.privateIPlb, tt.args.token)
		})
	}
}

func TestHAType_DeleteFirewall(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		firewallID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			tt.wantErr(t, obj.DeleteFirewall(tt.args.firewallID), fmt.Sprintf("DeleteFirewall(%v)", tt.args.firewallID))
		})
	}
}

func TestHAType_DeleteInstance(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		instanceID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			tt.wantErr(t, obj.DeleteInstance(tt.args.instanceID), fmt.Sprintf("DeleteInstance(%v)", tt.args.instanceID))
		})
	}
}

func TestHAType_DeleteInstances(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			tt.wantErr(t, obj.DeleteInstances(tt.args.logging), fmt.Sprintf("DeleteInstances(%v)", tt.args.logging))
		})
	}
}

func TestHAType_DeleteNetwork(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		networkID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			tt.wantErr(t, obj.DeleteNetwork(tt.args.networkID), fmt.Sprintf("DeleteNetwork(%v)", tt.args.networkID))
		})
	}
}

func TestHAType_DeleteNetworks(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			tt.wantErr(t, obj.DeleteNetworks(tt.args.logging), fmt.Sprintf("DeleteNetworks(%v)", tt.args.logging))
		})
	}
}

func TestHAType_DeleteSSHKeyPair(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			tt.wantErr(t, obj.DeleteSSHKeyPair(), fmt.Sprintf("DeleteSSHKeyPair()"))
		})
	}
}

func TestHAType_FetchKUBECONFIG(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		logging    logger.Logger
		instanceCP *civogo.Instance
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			got, err := obj.FetchKUBECONFIG(tt.args.logging, tt.args.instanceCP)
			if !tt.wantErr(t, err, fmt.Sprintf("FetchKUBECONFIG(%v, %v)", tt.args.logging, tt.args.instanceCP)) {
				return
			}
			assert.Equalf(t, tt.want, got, "FetchKUBECONFIG(%v, %v)", tt.args.logging, tt.args.instanceCP)
		})
	}
}

func TestHAType_GetInstance(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		instanceID string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantInst *civogo.Instance
		wantErr  assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			gotInst, err := obj.GetInstance(tt.args.instanceID)
			if !tt.wantErr(t, err, fmt.Sprintf("GetInstance(%v)", tt.args.instanceID)) {
				return
			}
			assert.Equalf(t, tt.wantInst, gotInst, "GetInstance(%v)", tt.args.instanceID)
		})
	}
}

func TestHAType_GetNetwork(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		networkName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantNet *civogo.Network
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			gotNet, err := obj.GetNetwork(tt.args.networkName)
			if !tt.wantErr(t, err, fmt.Sprintf("GetNetwork(%v)", tt.args.networkName)) {
				return
			}
			assert.Equalf(t, tt.wantNet, gotNet, "GetNetwork(%v)", tt.args.networkName)
		})
	}
}

func TestHAType_GetTokenFromCP_1(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		logging  logger.Logger
		instance *civogo.Instance
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
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			assert.Equalf(t, tt.want, obj.GetTokenFromCP_1(tt.args.logging, tt.args.instance), "GetTokenFromCP_1(%v, %v)", tt.args.logging, tt.args.instance)
		})
	}
}

func TestHAType_HelperExecNoOutputControlPlane(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		logging  logger.Logger
		publicIP string
		script   string
		fastMode bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			tt.wantErr(t, obj.HelperExecNoOutputControlPlane(tt.args.logging, tt.args.publicIP, tt.args.script, tt.args.fastMode), fmt.Sprintf("HelperExecNoOutputControlPlane(%v, %v, %v, %v)", tt.args.logging, tt.args.publicIP, tt.args.script, tt.args.fastMode))
		})
	}
}

func TestHAType_HelperExecOutputControlPlane(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		logging  logger.Logger
		publicIP string
		script   string
		fastMode bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			got, err := obj.HelperExecOutputControlPlane(tt.args.logging, tt.args.publicIP, tt.args.script, tt.args.fastMode)
			if !tt.wantErr(t, err, fmt.Sprintf("HelperExecOutputControlPlane(%v, %v, %v, %v)", tt.args.logging, tt.args.publicIP, tt.args.script, tt.args.fastMode)) {
				return
			}
			assert.Equalf(t, tt.want, got, "HelperExecOutputControlPlane(%v, %v, %v, %v)", tt.args.logging, tt.args.publicIP, tt.args.script, tt.args.fastMode)
		})
	}
}

func TestHAType_SaveKubeconfig(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		logging    logger.Logger
		kubeconfig string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			tt.wantErr(t, obj.SaveKubeconfig(tt.args.logging, tt.args.kubeconfig), fmt.Sprintf("SaveKubeconfig(%v, %v)", tt.args.logging, tt.args.kubeconfig))
		})
	}
}

func TestHAType_UploadSSHKey(t *testing.T) {
	type fields struct {
		Client        *civogo.Client
		ClusterName   string
		DiskImgID     string
		NetworkID     string
		NodeSize      string
		DBFirewallID  string
		LBFirewallID  string
		CPFirewallID  string
		WPFirewallID  string
		SSHID         string
		Configuration *JsonStore
		SSH_Payload   *utils.SSHPayload
	}
	type args struct {
		logging logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ha := &HAType{
				Client:        tt.fields.Client,
				ClusterName:   tt.fields.ClusterName,
				DiskImgID:     tt.fields.DiskImgID,
				NetworkID:     tt.fields.NetworkID,
				NodeSize:      tt.fields.NodeSize,
				DBFirewallID:  tt.fields.DBFirewallID,
				LBFirewallID:  tt.fields.LBFirewallID,
				CPFirewallID:  tt.fields.CPFirewallID,
				WPFirewallID:  tt.fields.WPFirewallID,
				SSHID:         tt.fields.SSHID,
				Configuration: tt.fields.Configuration,
				SSH_Payload:   tt.fields.SSH_Payload,
			}
			tt.wantErr(t, ha.UploadSSHKey(tt.args.logging), fmt.Sprintf("UploadSSHKey(%v)", tt.args.logging))
		})
	}
}

func TestJsonStore_ConfigWriterDBEndpoint(t *testing.T) {
	type fields struct {
		ClusterName string
		Region      string
		DBEndpoint  string
		ServerToken string
		SSHID       string
		InstanceIDs InstanceID
		NetworkIDs  NetworkID
	}
	type args struct {
		logging  logger.Logger
		endpoint string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &JsonStore{
				ClusterName: tt.fields.ClusterName,
				Region:      tt.fields.Region,
				DBEndpoint:  tt.fields.DBEndpoint,
				ServerToken: tt.fields.ServerToken,
				SSHID:       tt.fields.SSHID,
				InstanceIDs: tt.fields.InstanceIDs,
				NetworkIDs:  tt.fields.NetworkIDs,
			}
			tt.wantErr(t, config.ConfigWriterDBEndpoint(tt.args.logging, tt.args.endpoint), fmt.Sprintf("ConfigWriterDBEndpoint(%v, %v)", tt.args.logging, tt.args.endpoint))
		})
	}
}

func TestJsonStore_ConfigWriterFirewallControlPlaneNodes(t *testing.T) {
	type fields struct {
		ClusterName string
		Region      string
		DBEndpoint  string
		ServerToken string
		SSHID       string
		InstanceIDs InstanceID
		NetworkIDs  NetworkID
	}
	type args struct {
		logging logger.Logger
		fwID    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &JsonStore{
				ClusterName: tt.fields.ClusterName,
				Region:      tt.fields.Region,
				DBEndpoint:  tt.fields.DBEndpoint,
				ServerToken: tt.fields.ServerToken,
				SSHID:       tt.fields.SSHID,
				InstanceIDs: tt.fields.InstanceIDs,
				NetworkIDs:  tt.fields.NetworkIDs,
			}
			tt.wantErr(t, config.ConfigWriterFirewallControlPlaneNodes(tt.args.logging, tt.args.fwID), fmt.Sprintf("ConfigWriterFirewallControlPlaneNodes(%v, %v)", tt.args.logging, tt.args.fwID))
		})
	}
}

func TestJsonStore_ConfigWriterFirewallDatabaseNodes(t *testing.T) {
	type fields struct {
		ClusterName string
		Region      string
		DBEndpoint  string
		ServerToken string
		SSHID       string
		InstanceIDs InstanceID
		NetworkIDs  NetworkID
	}
	type args struct {
		logging logger.Logger
		fwID    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &JsonStore{
				ClusterName: tt.fields.ClusterName,
				Region:      tt.fields.Region,
				DBEndpoint:  tt.fields.DBEndpoint,
				ServerToken: tt.fields.ServerToken,
				SSHID:       tt.fields.SSHID,
				InstanceIDs: tt.fields.InstanceIDs,
				NetworkIDs:  tt.fields.NetworkIDs,
			}
			tt.wantErr(t, config.ConfigWriterFirewallDatabaseNodes(tt.args.logging, tt.args.fwID), fmt.Sprintf("ConfigWriterFirewallDatabaseNodes(%v, %v)", tt.args.logging, tt.args.fwID))
		})
	}
}

func TestJsonStore_ConfigWriterFirewallLoadBalancerNodes(t *testing.T) {
	type fields struct {
		ClusterName string
		Region      string
		DBEndpoint  string
		ServerToken string
		SSHID       string
		InstanceIDs InstanceID
		NetworkIDs  NetworkID
	}
	type args struct {
		logging logger.Logger
		fwID    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &JsonStore{
				ClusterName: tt.fields.ClusterName,
				Region:      tt.fields.Region,
				DBEndpoint:  tt.fields.DBEndpoint,
				ServerToken: tt.fields.ServerToken,
				SSHID:       tt.fields.SSHID,
				InstanceIDs: tt.fields.InstanceIDs,
				NetworkIDs:  tt.fields.NetworkIDs,
			}
			tt.wantErr(t, config.ConfigWriterFirewallLoadBalancerNodes(tt.args.logging, tt.args.fwID), fmt.Sprintf("ConfigWriterFirewallLoadBalancerNodes(%v, %v)", tt.args.logging, tt.args.fwID))
		})
	}
}

func TestJsonStore_ConfigWriterFirewallWorkerNodes(t *testing.T) {
	type fields struct {
		ClusterName string
		Region      string
		DBEndpoint  string
		ServerToken string
		SSHID       string
		InstanceIDs InstanceID
		NetworkIDs  NetworkID
	}
	type args struct {
		logging logger.Logger
		fwID    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &JsonStore{
				ClusterName: tt.fields.ClusterName,
				Region:      tt.fields.Region,
				DBEndpoint:  tt.fields.DBEndpoint,
				ServerToken: tt.fields.ServerToken,
				SSHID:       tt.fields.SSHID,
				InstanceIDs: tt.fields.InstanceIDs,
				NetworkIDs:  tt.fields.NetworkIDs,
			}
			tt.wantErr(t, config.ConfigWriterFirewallWorkerNodes(tt.args.logging, tt.args.fwID), fmt.Sprintf("ConfigWriterFirewallWorkerNodes(%v, %v)", tt.args.logging, tt.args.fwID))
		})
	}
}

func TestJsonStore_ConfigWriterInstanceControlPlaneNodes(t *testing.T) {
	type fields struct {
		ClusterName string
		Region      string
		DBEndpoint  string
		ServerToken string
		SSHID       string
		InstanceIDs InstanceID
		NetworkIDs  NetworkID
	}
	type args struct {
		logging    logger.Logger
		instanceID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &JsonStore{
				ClusterName: tt.fields.ClusterName,
				Region:      tt.fields.Region,
				DBEndpoint:  tt.fields.DBEndpoint,
				ServerToken: tt.fields.ServerToken,
				SSHID:       tt.fields.SSHID,
				InstanceIDs: tt.fields.InstanceIDs,
				NetworkIDs:  tt.fields.NetworkIDs,
			}
			tt.wantErr(t, config.ConfigWriterInstanceControlPlaneNodes(tt.args.logging, tt.args.instanceID), fmt.Sprintf("ConfigWriterInstanceControlPlaneNodes(%v, %v)", tt.args.logging, tt.args.instanceID))
		})
	}
}

func TestJsonStore_ConfigWriterInstanceDatabase(t *testing.T) {
	type fields struct {
		ClusterName string
		Region      string
		DBEndpoint  string
		ServerToken string
		SSHID       string
		InstanceIDs InstanceID
		NetworkIDs  NetworkID
	}
	type args struct {
		logging    logger.Logger
		instanceID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &JsonStore{
				ClusterName: tt.fields.ClusterName,
				Region:      tt.fields.Region,
				DBEndpoint:  tt.fields.DBEndpoint,
				ServerToken: tt.fields.ServerToken,
				SSHID:       tt.fields.SSHID,
				InstanceIDs: tt.fields.InstanceIDs,
				NetworkIDs:  tt.fields.NetworkIDs,
			}
			tt.wantErr(t, config.ConfigWriterInstanceDatabase(tt.args.logging, tt.args.instanceID), fmt.Sprintf("ConfigWriterInstanceDatabase(%v, %v)", tt.args.logging, tt.args.instanceID))
		})
	}
}

func TestJsonStore_ConfigWriterInstanceLoadBalancer(t *testing.T) {
	type fields struct {
		ClusterName string
		Region      string
		DBEndpoint  string
		ServerToken string
		SSHID       string
		InstanceIDs InstanceID
		NetworkIDs  NetworkID
	}
	type args struct {
		logging    logger.Logger
		instanceID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &JsonStore{
				ClusterName: tt.fields.ClusterName,
				Region:      tt.fields.Region,
				DBEndpoint:  tt.fields.DBEndpoint,
				ServerToken: tt.fields.ServerToken,
				SSHID:       tt.fields.SSHID,
				InstanceIDs: tt.fields.InstanceIDs,
				NetworkIDs:  tt.fields.NetworkIDs,
			}
			tt.wantErr(t, config.ConfigWriterInstanceLoadBalancer(tt.args.logging, tt.args.instanceID), fmt.Sprintf("ConfigWriterInstanceLoadBalancer(%v, %v)", tt.args.logging, tt.args.instanceID))
		})
	}
}

func TestJsonStore_ConfigWriterInstanceWorkerNodes(t *testing.T) {
	type fields struct {
		ClusterName string
		Region      string
		DBEndpoint  string
		ServerToken string
		SSHID       string
		InstanceIDs InstanceID
		NetworkIDs  NetworkID
	}
	type args struct {
		logging    logger.Logger
		instanceID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &JsonStore{
				ClusterName: tt.fields.ClusterName,
				Region:      tt.fields.Region,
				DBEndpoint:  tt.fields.DBEndpoint,
				ServerToken: tt.fields.ServerToken,
				SSHID:       tt.fields.SSHID,
				InstanceIDs: tt.fields.InstanceIDs,
				NetworkIDs:  tt.fields.NetworkIDs,
			}
			tt.wantErr(t, config.ConfigWriterInstanceWorkerNodes(tt.args.logging, tt.args.instanceID), fmt.Sprintf("ConfigWriterInstanceWorkerNodes(%v, %v)", tt.args.logging, tt.args.instanceID))
		})
	}
}

func TestJsonStore_ConfigWriterNetworkID(t *testing.T) {
	type fields struct {
		ClusterName string
		Region      string
		DBEndpoint  string
		ServerToken string
		SSHID       string
		InstanceIDs InstanceID
		NetworkIDs  NetworkID
	}
	type args struct {
		logging logger.Logger
		netID   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &JsonStore{
				ClusterName: tt.fields.ClusterName,
				Region:      tt.fields.Region,
				DBEndpoint:  tt.fields.DBEndpoint,
				ServerToken: tt.fields.ServerToken,
				SSHID:       tt.fields.SSHID,
				InstanceIDs: tt.fields.InstanceIDs,
				NetworkIDs:  tt.fields.NetworkIDs,
			}
			tt.wantErr(t, config.ConfigWriterNetworkID(tt.args.logging, tt.args.netID), fmt.Sprintf("ConfigWriterNetworkID(%v, %v)", tt.args.logging, tt.args.netID))
		})
	}
}

func TestJsonStore_ConfigWriterSSHID(t *testing.T) {
	type fields struct {
		ClusterName string
		Region      string
		DBEndpoint  string
		ServerToken string
		SSHID       string
		InstanceIDs InstanceID
		NetworkIDs  NetworkID
	}
	type args struct {
		logging    logger.Logger
		keypair_id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &JsonStore{
				ClusterName: tt.fields.ClusterName,
				Region:      tt.fields.Region,
				DBEndpoint:  tt.fields.DBEndpoint,
				ServerToken: tt.fields.ServerToken,
				SSHID:       tt.fields.SSHID,
				InstanceIDs: tt.fields.InstanceIDs,
				NetworkIDs:  tt.fields.NetworkIDs,
			}
			tt.wantErr(t, config.ConfigWriterSSHID(tt.args.logging, tt.args.keypair_id), fmt.Sprintf("ConfigWriterSSHID(%v, %v)", tt.args.logging, tt.args.keypair_id))
		})
	}
}

func TestJsonStore_ConfigWriterServerToken(t *testing.T) {
	type fields struct {
		ClusterName string
		Region      string
		DBEndpoint  string
		ServerToken string
		SSHID       string
		InstanceIDs InstanceID
		NetworkIDs  NetworkID
	}
	type args struct {
		logging logger.Logger
		token   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &JsonStore{
				ClusterName: tt.fields.ClusterName,
				Region:      tt.fields.Region,
				DBEndpoint:  tt.fields.DBEndpoint,
				ServerToken: tt.fields.ServerToken,
				SSHID:       tt.fields.SSHID,
				InstanceIDs: tt.fields.InstanceIDs,
				NetworkIDs:  tt.fields.NetworkIDs,
			}
			tt.wantErr(t, config.ConfigWriterServerToken(tt.args.logging, tt.args.token), fmt.Sprintf("ConfigWriterServerToken(%v, %v)", tt.args.logging, tt.args.token))
		})
	}
}

func Test_cleanup(t *testing.T) {
	type args struct {
		logging  logger.Logger
		provider CivoProvider
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, cleanup(tt.args.logging, tt.args.provider), fmt.Sprintf("cleanup(%v, %v)", tt.args.logging, tt.args.provider))
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
			assert.Equalf(t, tt.want, configLBscript(tt.args.controlPlaneIPs), "configLBscript(%v)", tt.args.controlPlaneIPs)
		})
	}
}

func Test_configWriterManaged(t *testing.T) {
	type args struct {
		logging    logger.Logger
		kubeconfig string
		clusterN   string
		region     string
		clusterID  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, configWriterManaged(tt.args.logging, tt.args.kubeconfig, tt.args.clusterN, tt.args.region, tt.args.clusterID), fmt.Sprintf("configWriterManaged(%v, %v, %v, %v, %v)", tt.args.logging, tt.args.kubeconfig, tt.args.clusterN, tt.args.region, tt.args.clusterID))
		})
	}
}

func Test_deleteClusterWithID(t *testing.T) {
	type args struct {
		logging    logger.Logger
		clusterID  string
		regionCode string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, deleteClusterWithID(tt.args.logging, tt.args.clusterID, tt.args.regionCode), fmt.Sprintf("deleteClusterWithID(%v, %v, %v)", tt.args.logging, tt.args.clusterID, tt.args.regionCode))
		})
	}
}

func Test_fetchAPIKey(t *testing.T) {
	type args struct {
		logger logger.Logger
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
			assert.Equalf(t, tt.want, fetchAPIKey(tt.args.logger), "fetchAPIKey(%v)", tt.args.logger)
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
			assert.Equalf(t, tt.want, generateDBPassword(tt.args.passwordLen), "generateDBPassword(%v)", tt.args.passwordLen)
		})
	}
}

func Test_haCreateClusterHandler(t *testing.T) {
	type args struct {
		logging  logger.Logger
		name     string
		region   string
		nodeSize string
		noCP     int
		noWP     int
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, haCreateClusterHandler(tt.args.logging, tt.args.name, tt.args.region, tt.args.nodeSize, tt.args.noCP, tt.args.noWP), fmt.Sprintf("haCreateClusterHandler(%v, %v, %v, %v, %v, %v)", tt.args.logging, tt.args.name, tt.args.region, tt.args.nodeSize, tt.args.noCP, tt.args.noWP))
		})
	}
}

func Test_haDeleteClusterHandler(t *testing.T) {
	type args struct {
		logging logger.Logger
		name    string
		region  string
		showMsg bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, haDeleteClusterHandler(tt.args.logging, tt.args.name, tt.args.region, tt.args.showMsg), fmt.Sprintf("haDeleteClusterHandler(%v, %v, %v, %v)", tt.args.logging, tt.args.name, tt.args.region, tt.args.showMsg))
		})
	}
}

func Test_isPresent(t *testing.T) {
	type args struct {
		offering    string
		clusterName string
		Region      string
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
			assert.Equalf(t, tt.want, isPresent(tt.args.offering, tt.args.clusterName, tt.args.Region), "isPresent(%v, %v, %v)", tt.args.offering, tt.args.clusterName, tt.args.Region)
		})
	}
}

func Test_isValidSizeHA(t *testing.T) {
	type args struct {
		size string
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
			assert.Equalf(t, tt.want, isValidSizeHA(tt.args.size), "isValidSizeHA(%v)", tt.args.size)
		})
	}
}

func Test_isValidSizeManaged(t *testing.T) {
	type args struct {
		size string
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
			assert.Equalf(t, tt.want, isValidSizeManaged(tt.args.size), "isValidSizeManaged(%v)", tt.args.size)
		})
	}
}

func Test_kubeconfigDeleter(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, kubeconfigDeleter(tt.args.path), fmt.Sprintf("kubeconfigDeleter(%v)", tt.args.path))
		})
	}
}

func Test_managedCreateClusterHandler(t *testing.T) {
	type args struct {
		logging    logger.Logger
		civoConfig CivoProvider
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, managedCreateClusterHandler(tt.args.logging, tt.args.civoConfig), fmt.Sprintf("managedCreateClusterHandler(%v, %v)", tt.args.logging, tt.args.civoConfig))
		})
	}
}

func Test_managedDeleteClusterHandler(t *testing.T) {
	type args struct {
		logging logger.Logger
		name    string
		region  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, managedDeleteClusterHandler(tt.args.logging, tt.args.name, tt.args.region, false), fmt.Sprintf("managedDeleteClusterHandler(%v, %v, %v)", tt.args.logging, tt.args.name, tt.args.region))
		})
	}
}

func Test_printer_Printer(t *testing.T) {
	type fields struct {
		ClusterName string
		Region      string
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
				ClusterName: tt.fields.ClusterName,
				Region:      tt.fields.Region,
			}
			p.Printer(log, tt.args.isHA, tt.args.operation)
		})
	}
}

func Test_saveConfig(t *testing.T) {
	type args struct {
		logging       logger.Logger
		clusterFolder string
		configStore   JsonStore
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, saveConfig(tt.args.logging, tt.args.clusterFolder, tt.args.configStore), fmt.Sprintf("saveConfig(%v, %v, %v)", tt.args.logging, tt.args.clusterFolder, tt.args.configStore))
		})
	}
}

func Test_saveConfigManaged(t *testing.T) {
	type args struct {
		logging       logger.Logger
		clusterFolder string
		configStore   ManagedConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, saveConfigManaged(tt.args.logging, tt.args.clusterFolder, tt.args.configStore), fmt.Sprintf("saveConfigManaged(%v, %v, %v)", tt.args.logging, tt.args.clusterFolder, tt.args.configStore))
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
			assert.Equalf(t, tt.want, scriptCP_n(tt.args.dbEndpoint, tt.args.privateIPlb, tt.args.token), "scriptCP_n(%v, %v, %v)", tt.args.dbEndpoint, tt.args.privateIPlb, tt.args.token)
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
			assert.Equalf(t, tt.want, scriptDB(tt.args.password), "scriptDB(%v)", tt.args.password)
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
			assert.Equalf(t, tt.want, scriptKUBECONFIG(), "scriptKUBECONFIG()")
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
			assert.Equalf(t, tt.want, scriptLB(), "scriptLB()")
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
			assert.Equalf(t, tt.want, scriptWP(tt.args.privateIPlb, tt.args.token), "scriptWP(%v, %v)", tt.args.privateIPlb, tt.args.token)
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
			assert.Equalf(t, tt.want, scriptWithCP_1(), "scriptWithCP_1()")
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
			assert.Equalf(t, tt.want, scriptWithoutCP_1(tt.args.dbEndpoint, tt.args.privateIPlb), "scriptWithoutCP_1(%v, %v)", tt.args.dbEndpoint, tt.args.privateIPlb)
		})
	}
}

func Test_validationOfArguments(t *testing.T) {
	type args struct {
		name   string
		region string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, validationOfArguments(tt.args.name, tt.args.region), fmt.Sprintf("validationOfArguments(%v, %v)", tt.args.name, tt.args.region))
		})
	}
}
