package cloud

import (
	azure_pkg "github.com/kubesimplify/ksctl/api/provider/azure"
	civo_pkg "github.com/kubesimplify/ksctl/api/provider/civo"
	"github.com/kubesimplify/ksctl/api/resources"
)

// create a type for controlelr
// which uses the controller.CreateHACluster(*KsctlClient) inside that

func HydrateCloud(client *resources.KsctlClient) {
	switch client.Metadata.Provider {
	case "civo":
		client.Cloud = civo_pkg.ReturnCivoStruct()
	case "azure":
		client.Cloud = azure_pkg.ReturnAzureStruct()
	}
}

func CreateHACluster(client *resources.KsctlClient) {
}
