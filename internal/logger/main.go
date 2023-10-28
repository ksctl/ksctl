package logger

import (
	"github.com/fatih/color"
	"io"
	"log/slog"
)

type Logger struct {
	logger     *slog.Logger
	moduleName string
}

func (l *Logger) SetModule(m string) {
	l.moduleName = "[" + m + "]"
}

func (l *Logger) New(verbose int, out io.Writer) {
	// LevelDebug Level = -4
	// LevelInfo  Level = 0
	// LevelWarn  Level = 4
	// LevelError Level = 8

	var ve slog.Level

	source := false
	if verbose < 0 {
		ve = slog.LevelDebug
		source = true
	} else if verbose < 4 {
		ve = slog.LevelInfo
	} else if verbose < 8 {
		ve = slog.LevelWarn
	} else {
		ve = slog.LevelError
	}

	if source {
		l.logger = slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
			AddSource: source,
			Level:     ve,
		}))
	} else {
		l.logger = slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{
			AddSource: false,
			Level:     ve,
		}))
	}
}

func (l *Logger) Print(msg string, args ...any) {
	l.logger.Info(l.moduleName+msg, args...)
}

func (l *Logger) Success(msg string, args ...any) {
	color.Set(color.FgGreen, color.Bold)
	defer color.Unset()
	l.logger.Info(l.moduleName+msg, args...)
}

func (l *Logger) Note(msg string, args ...any) {
	color.Set(color.FgBlue, color.Bold)
	defer color.Unset()
	l.logger.Info(l.moduleName+msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	defer color.Unset()
	l.logger.Debug(l.moduleName+msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	color.Set(color.FgHiRed, color.Bold)
	defer color.Unset()
	l.logger.Error(l.moduleName+msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	color.Set(color.FgYellow, color.Bold)
	defer color.Unset()
	l.logger.Warn(l.moduleName+msg, args...)
}
