package resources

import (
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	cloudController "github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
)

type LoggerFactory interface {
	Print(msg string, v ...any)

	Success(msg string, v ...any)

	Note(msg string, v ...any)

	Warn(msg string, v ...any)

	Error(msg string, v ...any)

	Debug(msg string, v ...any)

	SetPackageName(name string)

	// To be used by external logging
	ExternalLogHandler(msgType consts.CustomExternalLogLevel, message string)
	ExternalLogHandlerf(msgType consts.CustomExternalLogLevel, format string, args ...interface{})

	NewError(format string, v ...any) error

	Table(data []cloudController.AllClusterData)

	Box(title string, lines string)
}
