package cloud

import (
	"fmt"

	azure_pkg "github.com/kubesimplify/ksctl/api/provider/azure"
	civo_pkg "github.com/kubesimplify/ksctl/api/provider/civo"
	local_pkg "github.com/kubesimplify/ksctl/api/provider/local"
	"github.com/kubesimplify/ksctl/api/resources"
)

func HydrateCloud(client *resources.KsctlClient) {
	switch client.Metadata.Provider {
	case "civo":
		client.Cloud = civo_pkg.ReturnCivoStruct()
	case "azure":
		client.Cloud = azure_pkg.ReturnAzureStruct()
	case "local":
		client.Cloud = local_pkg.ReturnLocalStruct()
	default:
		panic("Invalid Cloud provider")
	}
	// call the init state for cloud providers
	_ = client.Cloud.InitState()
}

func CreateHACluster(client *resources.KsctlClient) error {
	_ = client.Cloud.NewNetwork(client.State)
	_ = client.Cloud.CreateUploadSSHKeyPair(client.State)
	fmt.Println("Firewall LB")
	_ = client.Cloud.NewFirewall(client.State)

	fmt.Println("Firewall DB")
	_ = client.Cloud.NewFirewall(client.State)
	fmt.Println("Firewall CP")
	_ = client.Cloud.NewFirewall(client.State)
	fmt.Println("Firewall WP")
	_ = client.Cloud.NewFirewall(client.State)

	fmt.Println("Loadbalancer VM")
	_ = client.Cloud.NewVM(client.State)
	for no := 0; no < int(client.Metadata.NoDS); no++ {
		fmt.Println("Datastore VM", no)
		_ = client.Cloud.NewVM(client.State)
	}

	for no := 0; no < int(client.Metadata.NoCP); no++ {
		fmt.Println("ControlPlane VM", no)
		_ = client.Cloud.NewVM(client.State)
	}

	for no := 0; no < int(client.Metadata.NoWP); no++ {
		fmt.Println("Workerplane VM", no)
		_ = client.Cloud.NewVM(client.State)
	}
	return nil
}

func CreateManagedCluster(client *resources.KsctlClient) error {
	_ = client.Cloud.NewManagedCluster(client.State)
	return nil
}
