package cloud

import (
	"fmt"
	"time"

	azure_pkg "github.com/kubesimplify/ksctl/api/provider/azure"
	civo_pkg "github.com/kubesimplify/ksctl/api/provider/civo"
	local_pkg "github.com/kubesimplify/ksctl/api/provider/local"
	"github.com/kubesimplify/ksctl/api/resources"
)

func HydrateCloud(client *resources.KsctlClient, operation string) {
	var err error
	switch client.Metadata.Provider {
	case "civo":
		client.Cloud, err = civo_pkg.ReturnCivoStruct(client.Metadata)
		if err != nil {
			panic("[cloud] " + err.Error())
		}
	case "azure":
		client.Cloud = azure_pkg.ReturnAzureStruct(client.Metadata)
	case "local":
		client.Cloud = local_pkg.ReturnLocalStruct(client.Metadata)
	default:
		panic("Invalid Cloud provider")
	}
	// call the init state for cloud providers
	err = client.Cloud.InitState(client.Storage, operation)
	if err != nil {
		panic("[cloud] " + err.Error())
	}

}

func DeleteHACluster(client *resources.KsctlClient) error {

	// TODO: ADD THE OTHER RESOURCE DESTRICTION

	// last one to delete is network
	err := client.Cloud.DelNetwork(client.Storage)
	if err != nil {
		return err
	}

	return nil
}

func CreateHACluster(client *resources.KsctlClient) error {
	err := client.Cloud.Name(client.ClusterName + "-net").NewNetwork(client.Storage)
	if err != nil {
		return err
	}
	_ = client.Cloud.Name("demo-ssh").CreateUploadSSHKeyPair(client.Storage)

	fmt.Println("Firewall LB")
	_ = client.Cloud.Name("demo-fw-lb").
		Role("loadbalancer").
		NewFirewall(client.Storage)

	fmt.Println("Firewall DB")
	_ = client.Cloud.Name("demo-fw-db").
		Role("datastore").
		NewFirewall(client.Storage)

	fmt.Println("Firewall CP")
	_ = client.Cloud.Name("demo-fw-cp").
		Role("controlplane").
		NewFirewall(client.Storage)

	fmt.Println("Firewall WP")
	_ = client.Cloud.Name("demo-fw-wp").
		Role("workerplane").
		NewFirewall(client.Storage)

	fmt.Println("Loadbalancer VM")
	_ = client.Cloud.Name("demo-vm-lb").
		Role("loadbalancer").
		VMType(client.LoadBalancerNodeType).
		Visibility(true).
		NewVM(client.Storage)

	for no := 0; no < client.Metadata.NoDS; no++ {
		_ = client.Cloud.Name(fmt.Sprintf("demo-vm-db-%d", no)).
			Role("datastore").
			VMType(client.DataStoreNodeType).
			Visibility(true).
			NewVM(client.Storage)
	}

	for no := 0; no < client.Metadata.NoCP; no++ {
		_ = client.Cloud.Name(fmt.Sprintf("demo-vm-cp-%d", no)).
			Role("controlplane").
			VMType(client.ControlPlaneNodeType).
			Visibility(true).
			NewVM(client.Storage)
	}

	for no := 0; no < client.Metadata.NoWP; no++ {
		_ = client.Cloud.Name(fmt.Sprintf("demo-vm-wp-%d", no)).
			Role("workerplane").
			VMType(client.WorkerPlaneNodeType).
			Visibility(true).
			NewVM(client.Storage)
	}
	return nil
}

// only for testing
// especially for testing the state mgt
func ManualPause() {
	fmt.Println("PAUSE FOR TESTING")
	time.Sleep(5 * time.Second)
}

func CreateManagedCluster(client *resources.KsctlClient) error {
	if err := client.Cloud.Name(client.Metadata.ClusterName + "-ksctl-managed-net").NewNetwork(client.Storage); err != nil {
		// need to verify wrt to other providers wrt network creation
		return err
	}
	if err := client.Cloud.Name(client.Metadata.ClusterName + "-ksctl-managed").
		VMType(client.Metadata.ManagedNodeType).
		NewManagedCluster(client.Storage); err != nil {
		return err
	}
	return nil
}

func DeleteManagedCluster(client *resources.KsctlClient) error {

	if err := client.Cloud.DelManagedCluster(client.Storage); err != nil {
		return err
	}

	// ManualPause()

	if err := client.Cloud.DelNetwork(client.Storage); err != nil {
		return err
	}
	client.Storage.Logger().Success("[cloud] Deleted the managed cluster")
	return nil
}
