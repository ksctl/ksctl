package controllers

import "github.com/kubesimplify/ksctl/pkg/resources"

type Controller interface {
	CreateManagedCluster(*resources.KsctlClient) error
	DeleteManagedCluster(*resources.KsctlClient) error

	SwitchCluster(*resources.KsctlClient) (*string, error)

	GetCluster(*resources.KsctlClient) error

	Credentials(*resources.KsctlClient) error

	CreateHACluster(*resources.KsctlClient) error
	DeleteHACluster(*resources.KsctlClient) error

	AddWorkerPlaneNode(*resources.KsctlClient) error
	DelWorkerPlaneNode(*resources.KsctlClient) error
}
