package cloud

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/provider/civo"
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

func NewController(b *cloud.ClientBuilder) {
	// TODO: Which one to call controller will decide
	var abcd cloud.ControllerInterface = civo.WrapCloudControllerBuilder(b)
	fmt.Printf("\tReq for HA: %v\n\n", b.IsHA)
	abcd.CreateHACluster()
	abcd.CreateManagedCluster()
	abcd.DestroyHACluster()
	abcd.DestroyManagedCluster()
}

func WrapEngineBuilder(b *resources.Builder) *cloud.ClientBuilder {
	api := (*cloud.ClientBuilder)(b)
	return api
}
