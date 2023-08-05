package controllers

import "github.com/kubesimplify/ksctl/api/resources"

type Controller interface {
	CreateManagedCluster(*resources.KsctlClient)
	DeleteManagedCluster(*resources.KsctlClient)

	SwitchCluster()

	GetCluster(*resources.KsctlClient)

	Credentials(*resources.KsctlClient)

	CreateHACluster(*resources.KsctlClient)
	DeleteHACluster(*resources.KsctlClient)
}
