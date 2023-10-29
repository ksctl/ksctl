package resources

import (
	cloudController "github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
)

type LoggerFactory interface {
	Print(msg string, v ...any)

	Success(msg string, v ...any)

	Note(msg string, v ...any)

	Warn(msg string, v ...any)

	Error(msg string, v ...any)

	Debug(msg string, v ...any)

	SetPackageName(name string)

	NewError(format string, v ...any) error

	Table(data []cloudController.AllClusterData)
}
