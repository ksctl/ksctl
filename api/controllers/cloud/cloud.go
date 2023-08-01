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
	_ = client.Cloud.Name("demo-net").NewNetwork(client.State)
	_ = client.Cloud.Name("demo-ssh").CreateUploadSSHKeyPair(client.State)

	fmt.Println("Firewall LB")
	_ = client.Cloud.Name("demo-fw-lb").
		Role("loadbalancer").
		VMType("abcd-.t3mico").
		NewFirewall(client.State)

	fmt.Println("Firewall DB")
	_ = client.Cloud.Name("demo-fw-db").
		Role("datastore").
		VMType("abcd-.t3medium").
		NewFirewall(client.State)

	fmt.Println("Firewall CP")
	_ = client.Cloud.Name("demo-fw-cp").
		Role("controlplane").
		VMType("abcd-.t3large").
		NewFirewall(client.State)

	fmt.Println("Firewall WP")
	_ = client.Cloud.Name("demo-fw-wp").
		Role("workerplane").
		VMType("abcd-.t3large").
		NewFirewall(client.State)

	fmt.Println("Loadbalancer VM")
	_ = client.Cloud.Name("demo-vm-lb").
		Role("loadbalancer").
		VMType("abcd-.t3mico").
		Visibility(true).
		NewVM(client.State)

	for no := 0; no < int(client.Metadata.NoDS); no++ {
		_ = client.Cloud.Name(fmt.Sprintf("demo-vm-db-%d", no)).
			Role("datastore").
			VMType("abcd-.t3medium").
			Visibility(true).
			NewVM(client.State)
	}

	for no := 0; no < int(client.Metadata.NoCP); no++ {
		_ = client.Cloud.Name(fmt.Sprintf("demo-vm-cp-%d", no)).
			Role("controlplane").
			VMType("abcd-.t3mico").
			Visibility(true).
			NewVM(client.State)
	}

	for no := 0; no < int(client.Metadata.NoWP); no++ {
		_ = client.Cloud.Name(fmt.Sprintf("demo-vm-wp-%d", no)).
			Role("workerplane").
			VMType("abcd-.t3mico").
			Visibility(true).
			NewVM(client.State)
	}
	return nil
}

func CreateManagedCluster(client *resources.KsctlClient) error {
	_ = client.Cloud.Name(client.Metadata.ClusterName).NewManagedCluster(client.State)
	return nil
}
