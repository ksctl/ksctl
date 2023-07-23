package kubernetes

import (
	"github.com/kubesimplify/ksctl/api/resources"
	"github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
)

type ClientBuilder resources.Builder

type ClientStateMgt resources.StateManagementInfrastructure

type ControllerInterface interface {
	SetupLoadBalancer()
	GetServerToken() (string, error)
	GetKubeconfig() (string, error)
	SetupDatastore() (string, error)
	InitializeMasterControlPlane() error
	SetupWorkerplane() (string, error)
	JoinWorkerplane() (string, error)
	JoinControlplane() (string, error)
	JoinDatastore() (string, error)
	HydrateCloudState(cloud.CloudResourceState)
}
