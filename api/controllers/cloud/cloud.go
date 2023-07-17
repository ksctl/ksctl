package cloud

import "github.com/kubesimplify/ksctl/api/resources"

const (
	CREATE = 0
	DELETE = 1
	GET    = 2
)

type ClientBuilder resources.Builder

type ClientStateMgt resources.StateManagementInfrastructure

// func (client *ClientBuilder) Controller(operation uint8) {
// 	switch operation {
// 	case CREATE:
// 		// createCluster(client)
// 	case DELETE:
// 		// deleteCluster(client)
// 	}
// }
//
// func (client *ClientBuilder) APIhandler() {
// 	// controller
// 	client.Controller(CREATE)
// }

type ControllerInterface interface {
	CreateHACluster()
	CreateManagedCluster()

	DestroyHACluster()
	DestroyManagedCluster()
}
