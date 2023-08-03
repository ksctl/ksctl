package cloud

import (
	"fmt"
	"log"

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
			log.Fatal(err)
		}
	case "azure":
		client.Cloud = azure_pkg.ReturnAzureStruct(client.Metadata)
	case "local":
		client.Cloud = local_pkg.ReturnLocalStruct(client.Metadata)
	default:
		panic("Invalid Cloud provider")
	}
	// call the init state for cloud providers
	_ = client.Cloud.InitState(client.State, operation)
}

func DeleteHACluster(client *resources.KsctlClient) error {

	// add more

	// last one t delete is network

	err := client.Cloud.Name(client.ClusterName + "-net").DelNetwork(client.State)
	if err != nil {
		return err
	}

	return nil
}

func CreateHACluster(client *resources.KsctlClient) error {
	err := client.Cloud.Name(client.ClusterName + "-net").NewNetwork(client.State)
	if err != nil {
		return err
	}
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
