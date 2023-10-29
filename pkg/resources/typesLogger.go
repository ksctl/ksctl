package resources

import (
	cloudController "github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
)

type LoggerFactory interface {
	Print(string, ...any)

	Success(string, ...any)

	Note(string, ...any)

	Warn(string, ...any)

	Error(string, ...any)

	Debug(string, ...any)

	SetPackageName(string)

	Table(data []cloudController.AllClusterData)
}
