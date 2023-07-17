package cloud

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/provider/azure"
	"github.com/kubesimplify/ksctl/api/provider/civo"
	"github.com/kubesimplify/ksctl/api/provider/local"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

func NewController(b *cloud.ClientBuilder) {
	// TODO: Which one to call controller will decide
	var abcd cloud.ControllerInterface
	switch b.Provider {
	case "civo":
		abcd = civo.WrapCloudControllerBuilder(b)
	case "azure":
		abcd = azure.WrapCloudControllerBuilder(b)
	case "local":
		abcd = local.WrapCloudControllerBuilder(b)
	}
	fmt.Printf("\tReq for HA: %v\n\n", b.IsHA)
	// abcd.CreateHACluster() // for local it will panic
	abcd.CreateManagedCluster()
	// abcd.DestroyHACluster() // for local it will panic
	abcd.DestroyManagedCluster()
}

func WrapCloudEngineBuilder(b *resources.Builder) *cloud.ClientBuilder {
	api := (*cloud.ClientBuilder)(b)
	return api
}
