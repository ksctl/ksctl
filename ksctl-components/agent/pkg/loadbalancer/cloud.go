package loadbalancer

import "github.com/ksctl/ksctl/pkg/helpers/consts"

type CloudLB interface {
	Create() (*string, error)
	Delete() error
}

func InitProvider(cloud consts.KsctlCloud) CloudLB {
	switch cloud {
	case consts.CloudCivo:
		return &civoLB{}
	default:
		return nil
	}
}
