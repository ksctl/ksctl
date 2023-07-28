package cloud

import (
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
	netErr := client.Cloud.NewNetwork(client.State)
	if netErr != nil {
		return netErr
	}

	return nil
}
