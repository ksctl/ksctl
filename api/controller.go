package api

import (
	"github.com/kubesimplify/ksctl/api/controllers/cloud"
	"github.com/kubesimplify/ksctl/api/provider/civo"
)

func NewController(b *cloud.ClientBuilder) {

	var abcd cloud.ControllerInterface = civo.WrapCloudControllerBuilder(b)
	abcd.CreateHACluster()
	abcd.CreateManagedCluster()
	abcd.DestroyHACluster()
	abcd.DestroyManagedCluster()
}
