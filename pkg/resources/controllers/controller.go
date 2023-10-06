package controllers

import "github.com/kubesimplify/ksctl/pkg/resources"

type Controller interface {
	CreateManagedCluster(*resources.KsctlClient) (string, error)
	DeleteManagedCluster(*resources.KsctlClient) (string, error)

	SwitchCluster(*resources.KsctlClient) (string, error)

	GetCluster(*resources.KsctlClient) (string, error)

	Credentials(*resources.KsctlClient) (string, error)

	CreateHACluster(*resources.KsctlClient) (string, error)
	DeleteHACluster(*resources.KsctlClient) (string, error)

	AddWorkerPlaneNode(*resources.KsctlClient) (string, error)
	DelWorkerPlaneNode(*resources.KsctlClient) (string, error)
}
