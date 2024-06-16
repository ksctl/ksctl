package types

import (
	"context"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	cloudController "github.com/ksctl/ksctl/pkg/types/controllers/cloud"
)

type LoggerFactory interface {
	Print(ctx context.Context, msg string, v ...any)

	Success(ctx context.Context, msg string, v ...any)

	Note(ctx context.Context, msg string, v ...any)

	Warn(ctx context.Context, msg string, v ...any)

	Error(msg string, v ...any)

	Debug(ctx context.Context, msg string, v ...any)

	ExternalLogHandler(ctx context.Context, msgType consts.CustomExternalLogLevel, message string)
	ExternalLogHandlerf(ctx context.Context, msgType consts.CustomExternalLogLevel, format string, args ...interface{})

	NewError(ctx context.Context, msg string, v ...any) error

	Table(ctx context.Context, operation consts.LogClusterDetail, data []cloudController.AllClusterData)

	Box(ctx context.Context, title string, lines string)
}
