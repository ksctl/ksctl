package controllers

import (
	cloudController "github.com/kubesimplify/ksctl/api/controllers/cloud"
	k8sController "github.com/kubesimplify/ksctl/api/controllers/kubernetes"
	"github.com/kubesimplify/ksctl/api/resources"
)

func NewController(client *resources.Builder) {
	ksctlCloudAPI := cloudController.WrapCloudEngineBuilder(client)
	abcd := cloudController.NewController(ksctlCloudAPI)
	reqForK8sDistro := abcd.FetchState()

	ksctlK8sAPI := k8sController.WrapK8sEngineBuilder(client)
	k8sController.NewController(ksctlK8sAPI, reqForK8sDistro)
}
