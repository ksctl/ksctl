package controllers

import "github.com/kubesimplify/ksctl/api/resources"

type Controller interface {
	CreateManagedCluster(*resources.KsctlClient)
	DeleteManagedCluster(*resources.KsctlClient)

	SwitchCluster()

	GetCluster()

	CreateHACluster(*resources.KsctlClient)
	DeleteHACluster(*resources.KsctlClient)
}
