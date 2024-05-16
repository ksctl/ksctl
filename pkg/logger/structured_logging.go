package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/fatih/color"
	cloudController "github.com/ksctl/ksctl/pkg/types/controllers/cloud"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

type StructuredLog struct {
	logger *slog.Logger
}

const (
	limitCol = 80
)

func (l *StructuredLog) ExternalLogHandler(ctx context.Context, msgType consts.CustomExternalLogLevel, message string) {
	_m := ""
	switch msgType {
	case consts.LOG_DEBUG:
		_m = "DEBUG"
	case consts.LOG_ERROR:
		_m = "ERROR"
	case consts.LOG_SUCCESS:
		_m = "SUCCESS"
	case consts.LOG_WARNING:
		_m = "WARN"
	default:
		_m = "INFO"
	}
	l.logger.Info(message, "component", ctx.Value(consts.ContextModuleNameKey), "msgType", _m)
}

func (l *StructuredLog) ExternalLogHandlerf(ctx context.Context, msgType consts.CustomExternalLogLevel, format string, args ...interface{}) {
	_m := ""
	switch msgType {
	case consts.LOG_DEBUG:
		_m = "DEBUG"
	case consts.LOG_ERROR:
		_m = "ERROR"
	case consts.LOG_SUCCESS:
		_m = "SUCCESS"
	case consts.LOG_WARNING:
		_m = "WARN"
	default:
		_m = "INFO"
	}
	l.logger.Info(fmt.Sprintf(format, args...), "component", ctx.Value(consts.ContextModuleNameKey), "msgType", _m)
}

func newLogger(out io.Writer, ver slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: ver,
	}))
}

func NewStructuredLogger(verbose int, out io.Writer) *StructuredLog {
	// LevelDebug Level = -4
	// LevelInfo  Level = 0
	// LevelWarn  Level = 4
	// LevelError Level = 8

	var ve slog.Level

	if verbose < 0 {
		ve = slog.LevelDebug
	} else if verbose < 4 {
		ve = slog.LevelInfo
	} else if verbose < 8 {
		ve = slog.LevelWarn
	} else {
		ve = slog.LevelError
	}

	return &StructuredLog{logger: newLogger(out, ve)}
}

func (l *StructuredLog) Print(ctx context.Context, msg string, args ...any) {
	args = append([]any{"component", ctx.Value(consts.ContextModuleNameKey)}, args...)
	l.logger.Info(msg, args...)
}

func (l *StructuredLog) Success(ctx context.Context, msg string, args ...any) {
	color.Set(color.FgGreen, color.Bold)
	defer color.Unset()
	args = append([]any{"component", ctx.Value(consts.ContextModuleNameKey), "msgType", "SUCCESS"}, args...)
	l.logger.Info(msg, args...)
}

func (l *StructuredLog) Note(ctx context.Context, msg string, args ...any) {
	color.Set(color.FgBlue, color.Bold)
	defer color.Unset()
	args = append([]any{"component", ctx.Value(consts.ContextModuleNameKey), "msgType", "NOTE"}, args...)
	l.logger.Info(msg, args...)
}

func (l *StructuredLog) Debug(ctx context.Context, msg string, args ...any) {
	defer color.Unset()
	args = append([]any{"component", ctx.Value(consts.ContextModuleNameKey)}, args...)
	l.logger.Debug(msg, args...)
}

func (l *StructuredLog) Error(ctx context.Context, msg string, args ...any) {
	color.Set(color.FgHiRed, color.Bold)
	defer color.Unset()

	l.logger.Error(msg, args...)
}

func (l *StructuredLog) NewError(ctx context.Context, format string, args ...any) error {
	args = append([]any{"component", ctx.Value(consts.ContextModuleNameKey)}, args...)
	return fmt.Errorf(format, args...)
}

func (l *StructuredLog) Warn(ctx context.Context, msg string, args ...any) {
	color.Set(color.FgYellow, color.Bold)
	defer color.Unset()
	args = append([]any{"component", ctx.Value(consts.ContextModuleNameKey), "msgType", "WARN"}, args...)
	l.logger.Warn(msg, args...)
}

func (l *StructuredLog) Table(ctx context.Context, data []cloudController.AllClusterData) {

	type nodeSchema struct {
		Cp int
		Wp int
		Lb int
		Ds int
	}
	type tableSchema struct {
		Name         string
		Provider     string
		Region       string
		Nodes        nodeSchema
		Type         string
		K8sBootstrap string
		K8sVersion   string
	}

	vals := []tableSchema{}

	for _, row := range data {
		_val := tableSchema{}
		_val.Name = row.Name
		_val.Provider = string(row.Provider)
		_val.Region = row.Region
		_val.Type = string(row.Type)
		_val.K8sBootstrap = string(row.K8sDistro)
		_val.K8sVersion = row.K8sVersion

		if row.Type == "ha" {
			_val.Nodes.Cp = row.NoCP
			_val.Nodes.Wp = row.NoWP
			_val.Nodes.Ds = row.NoDS
			_val.Nodes.Lb = 1
		} else {
			_val.Nodes.Wp = row.NoMgt
		}

		vals = append(vals, _val)
	}
	l.Success(ctx, "table content", "data", vals)
}

func (l *StructuredLog) Box(ctx context.Context, title string, lines string) {

	l.Print(ctx, title, "details", addLineTerminationForLongStrings(lines))
}
