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
	case consts.LogDebug:
		_m = "DEBUG"
	case consts.LogError:
		_m = "ERROR"
	case consts.LogSuccess:
		_m = "SUCCESS"
	case consts.LogWarning:
		_m = "WARN"
	default:
		_m = "INFO"
	}
	l.logger.Info(message, "component", ctx.Value(consts.KsctlModuleNameKey), "msgType", _m)
}

func (l *StructuredLog) ExternalLogHandlerf(ctx context.Context, msgType consts.CustomExternalLogLevel, format string, args ...interface{}) {
	_m := ""
	switch msgType {
	case consts.LogDebug:
		_m = "DEBUG"
	case consts.LogError:
		_m = "ERROR"
	case consts.LogSuccess:
		_m = "SUCCESS"
	case consts.LogWarning:
		_m = "WARN"
	default:
		_m = "INFO"
	}
	l.logger.Info(fmt.Sprintf(format, args...), "component", ctx.Value(consts.KsctlModuleNameKey), "msgType", _m)
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
	args = append([]any{"component", ctx.Value(consts.KsctlModuleNameKey)}, args...)
	l.logger.Info(msg, args...)
}

func (l *StructuredLog) Success(ctx context.Context, msg string, args ...any) {
	color.Set(color.FgGreen, color.Bold)
	defer color.Unset()
	args = append([]any{"component", ctx.Value(consts.KsctlModuleNameKey), "msgType", "SUCCESS"}, args...)
	l.logger.Info(msg, args...)
}

func (l *StructuredLog) Note(ctx context.Context, msg string, args ...any) {
	color.Set(color.FgBlue, color.Bold)
	defer color.Unset()
	args = append([]any{"component", ctx.Value(consts.KsctlModuleNameKey), "msgType", "NOTE"}, args...)
	l.logger.Info(msg, args...)
}

func (l *StructuredLog) Debug(ctx context.Context, msg string, args ...any) {
	defer color.Unset()
	args = append([]any{"component", ctx.Value(consts.KsctlModuleNameKey)}, args...)
	l.logger.Debug(msg, args...)
}

func (l *StructuredLog) Error(ctx context.Context, msg string, args ...any) {
	color.Set(color.FgHiRed, color.Bold)
	defer color.Unset()

	l.logger.Error(msg, args...)
}

func (l *StructuredLog) NewError(ctx context.Context, format string, args ...any) error {
	args = append([]any{"component", ctx.Value(consts.KsctlModuleNameKey)}, args...)
	return fmt.Errorf(format, args...)
}

func (l *StructuredLog) Warn(ctx context.Context, msg string, args ...any) {
	color.Set(color.FgYellow, color.Bold)
	defer color.Unset()
	args = append([]any{"component", ctx.Value(consts.KsctlModuleNameKey), "msgType", "WARN"}, args...)
	l.logger.Warn(msg, args...)
}

func (l *StructuredLog) Table(ctx context.Context, data []cloudController.AllClusterData) {

	l.Success(ctx, "table content", "data", data)
}

func (l *StructuredLog) Box(ctx context.Context, title string, lines string) {

	l.Print(ctx, title, "details", addLineTerminationForLongStrings(lines))
}
