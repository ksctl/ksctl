package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/kubesimplify/ksctl/pkg/utils/consts"
	"go.mongodb.org/mongo-driver/bson"
)

var storage ConfigurationStore

func TestMain(m *testing.M) {

	fmt.Println("Init")
	_ = os.Setenv("MONGODB_HOSTNAME", "cluster0.r8uly4m.mongodb.net")
	_ = os.Setenv("MONGODB_USER", "dipankar")
	_ = os.Setenv("MONGODB_PASSWORD", "1234")

	hostname := os.Getenv("MONGODB_HOSTNAME")
	user := os.Getenv("MONGODB_USER")
	password := os.Getenv("MONGODB_PASSWORD")

	var err error
	storage, err = NewClient(Options{
		Hostname: hostname,
		Username: user,
		Password: password})
	if err != nil {
		panic(err)
	}

	exitVal := m.Run()

	defer func() {
		if err := storage.Disconnect(); err != nil {
			panic(err)
		}
	}()

	os.Exit(exitVal)
}

func TestListDatabases(t *testing.T) {
	if databases, err := storage.ListDatabases(); err != nil {
		return
	} else {
		fmt.Println(databases)
	}
}

func TestWrite(t *testing.T) {
	state := StorageDocument{ClusterType: "managed", Region: "eastus", ClusterName: "demo"}
	state.CloudInfra = &InfrastructureState{Civo: &CivoState{}}

	if err := storage.Write("civo", state); err != nil {
		t.Fatalf("Unable to write: %v", err)
	}

	state.ClusterType = "ha"
	state.Region = "eastus1"
	if err := storage.Write("civo", state); err != nil {
		t.Fatalf("Unable to write: %v", err)
	}

	state.ClusterType = "managed"
	state.Region = "westus"
	state.ClusterName = "demo1"
	if err := storage.Write("civo", StorageDocument{ClusterType: "managed", Region: "westus", ClusterName: "demo1"}); err != nil {
		t.Fatalf("Unable to write: %v", err)
	}
}

func TestGetAll(t *testing.T) {
	data, err := storage.GetAllClusters("civo", bson.M{"cluster_type": "ha"})

	if err != nil {
		t.Fatalf("unable to get all the clusters %v", err)
	}
	fmt.Printf("[[ CIVO HA ]]%+v\n", data)

	data, err = storage.GetAllClusters("civo", bson.M{"cluster_type": "managed"})

	if err != nil {
		t.Fatalf("unable to get all the clusters %v", err)
	}
	fmt.Printf("[[ CIVO Managed ]]%+v\n", data)
}

func TestRead(t *testing.T) {

	if data, err := storage.ReadOne("civo", "eastus", "demo", "managed"); err != nil {
		t.Fatalf("Unable to Read: %v", err)
	} else {
		fmt.Printf("%+v\n", data)
	}
}

func TestPresent(t *testing.T) {
	if ok := storage.IsPresent("civo", "eastus", "demo", "managed"); !ok {
		t.Fatalf("expected that civo, eastus, demo, managed is present but got its false")
	}

	if ok := storage.IsPresent("azure", "eastus", "demo", "managed"); ok {
		t.Fatalf("expected that azure, eastus, demo, managed is absent but got its true")
	}
}

func TestDelete(t *testing.T) {

	if err := storage.DeleteOne("civo", "eastus", "demo", "managed"); err != nil {
		t.Fatalf("Unable to Delete: %v", err)
	}
}

func TestDeleteAll(t *testing.T) {

	if err := storage.DeleteAllInCloud("civo"); err != nil {
		t.Fatalf("Unable to Delete: %v", err)
	}
}

func TestEntireMongoStore(t *testing.T) {
	createAzure(t)
}

var (
	CLUSTER_NAME string = "demo-poc"
	REGION       string = "ap-south-1"
	CLOUD        string = "azure"
	K8S          string = "k3s"
)

func createAzure(t *testing.T) {
	if err := createManaged(StorageDocument{}); err != nil {
		t.Fatalf("failed to create the managed cluster: %v", err)
	}
	if err := createHA(StorageDocument{}); err != nil {
		t.Fatalf("failed to create the ha cluster: %v", err)
	}

	if err := deleteHA(); err != nil {
		t.Fatalf("failed to delete the ha cluster: %v", err)
	}
	if err := deleteManaged(); err != nil {
		t.Fatalf("failed to delete the managed cluster: %v", err)
	}
}

// azzuming we know its azure and k3s
func createManaged(state StorageDocument) error {
	fmt.Println("@@@@ MANAGED CREATE @@@@")

	state.CloudInfra = &InfrastructureState{Azure: &AzureState{}}

	state.ClusterName = CLUSTER_NAME
	state.Region = REGION
	state.ClusterType = string(consts.ClusterTypeMang)

	if err := createCloudPkg(state); err != nil {
		return err
	}

	return nil
}

func createHA(state StorageDocument) error {
	fmt.Println("@@@@ HA CREATE @@@@")

	state.CloudInfra = &InfrastructureState{Azure: &AzureState{}}
	state.BootStrapConfig = &KubernetesBootstrapState{K3s: &K3sBootstrapState{}}

	state.ClusterName = CLUSTER_NAME
	state.Region = REGION
	state.ClusterType = string(consts.ClusterTypeHa)

	if err := createCloudPkg(state); err != nil {
		return err
	}
	if err := createKubernetesPkg(state); err != nil {
		return err
	}

	return nil
}

func createCloudPkg(state StorageDocument) error {

	state.CloudInfra.Azure.IsCompleted = true
	state.CloudInfra.Azure.ResourceGroupName = "demo"

	if state.ClusterType == string(consts.ClusterTypeMang) {
		state.CloudInfra.Azure.ManagedClusterName = "demo"
		state.CloudInfra.Azure.NoManagedNodes = 10

		// simulating kubeconfig save
		state.ClusterKubeConfig = "config"
	} else {
		state.CloudInfra.Azure.SubnetName = "dscsdc"
		state.CloudInfra.Azure.SSHKeyName = "demo"
		state.CloudInfra.Azure.VirtualNetworkID = "demo"
	}

	return storage.Write(CLOUD, state)
}

func createKubernetesPkg(state StorageDocument) error {
	// its only invoked when ha is called

	state.BootStrapConfig.K3s.SSHInfo.UserName = "root"
	state.BootStrapConfig.K3s.K3sToken = "K3sscdcdwscsdfvcfdsv"
	state.BootStrapConfig.K3s.DataStoreEndPoint = "mysqlcxdscscsc"

	// simulating kubeconfig save
	state.ClusterKubeConfig = "config"

	return storage.Write(CLOUD, state)
}

func deleteManaged() error {
	fmt.Println("@@@@ MANAGED DELETE @@@@")
	state, err := storage.ReadOne(CLOUD, REGION, CLUSTER_NAME, string(consts.ClusterTypeMang))
	if err != nil {
		return err
	}

	raw, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		return err
	}
	fmt.Println(string(raw))

	return storage.DeleteOne(CLOUD, REGION, CLUSTER_NAME, string(consts.ClusterTypeMang))
}

func deleteHA() error {
	fmt.Println("@@@@ HA DELETE @@@@")
	state, err := storage.ReadOne(CLOUD, REGION, CLUSTER_NAME, string(consts.ClusterTypeHa))
	if err != nil {
		return err
	}

	raw, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		return err
	}
	fmt.Println(string(raw))

	state.BootStrapConfig.K3s = nil
	if err := storage.Write(CLOUD, state); err != nil {
		return err
	}

	state, err = storage.ReadOne(CLOUD, REGION, CLUSTER_NAME, string(consts.ClusterTypeHa))
	if err != nil {
		return err
	}

	raw, err = json.MarshalIndent(state, "", " ")
	if err != nil {
		return err
	}
	fmt.Println(string(raw))

	return storage.DeleteOne(CLOUD, REGION, CLUSTER_NAME, string(consts.ClusterTypeHa))
}
