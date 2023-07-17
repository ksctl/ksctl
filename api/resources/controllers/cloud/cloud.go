package cloud

import "github.com/kubesimplify/ksctl/api/resources"

const (
	CREATE = 0
	DELETE = 1
	GET    = 2
)

type ClientBuilder resources.Builder

type ClientStateMgt resources.StateManagementInfrastructure

type ControllerInterface interface {
	CreateHACluster()
	CreateManagedCluster()

	DestroyHACluster()
	DestroyManagedCluster()
}
