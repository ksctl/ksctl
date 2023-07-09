package cli

import "github.com/kubesimplify/ksctl/api/resources"

type Builder struct {
	Client resources.CloudInfrastructure
}

type CobraCmd struct {
	ClusterName string
	Region      string
	Client      Builder
	Version     string
}
