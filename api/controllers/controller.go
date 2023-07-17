package controllers

import (
	cloudController "github.com/kubesimplify/ksctl/api/controllers/cloud"
	k8sController "github.com/kubesimplify/ksctl/api/controllers/kubernetes"
	"github.com/kubesimplify/ksctl/api/resources"
)

func NewController(client *resources.Builder) {
	ksctlCloudAPI := cloudController.WrapCloudEngineBuilder(client)
	cloudController.NewController(ksctlCloudAPI)

	ksctlK8sAPI := k8sController.WrapK8sEngineBuilder(client)
	k8sController.NewController(ksctlK8sAPI)
}
