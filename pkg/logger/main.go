package logger

import cloudController "github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"

type Logger struct {
	Verbose bool
}

type LogFactory interface {
	// will accept a string to be highlignted
	Success(...string)
	Warn(...string)
	Print(...string)
	Err(...string)
	Note(...string)
	Table([]cloudController.AllClusterData)
}
