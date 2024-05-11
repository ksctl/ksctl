package controllers

import (
	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

type Controller interface {
	CreateManagedCluster() error
	DeleteManagedCluster() error

	SwitchCluster() (*string, error)
	Applications(consts.KsctlOperation) error
	GetCluster() error

	Credentials() error

	CreateHACluster() error
	DeleteHACluster() error

	AddWorkerPlaneNode() error
	DelWorkerPlaneNode() error
}
