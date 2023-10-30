package logger

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/fatih/color"
	"github.com/kubesimplify/ksctl/pkg/resources"
	cloudController "github.com/kubesimplify/ksctl/pkg/resources/controllers/cloud"
	"github.com/rodaine/table"
)

type Logger struct {
	logger     *slog.Logger
	moduleName string
}

func (l *Logger) SetPackageName(m string) {
	l.moduleName = m
}

func newLogger(out io.Writer, ver slog.Level, debug bool) *slog.Logger {
	if debug {
		return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
			AddSource: true,
			Level:     ver,
		}))
	}
	return slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{
		AddSource: false,
		Level:     ver,
	}))
}

func NewDefaultLogger(verbose int, out io.Writer) resources.LoggerFactory {
	// LevelDebug Level = -4
	// LevelInfo  Level = 0
	// LevelWarn  Level = 4
	// LevelError Level = 8

	var ve slog.Level

	if verbose < 0 {
		ve = slog.LevelDebug

		return &Logger{logger: newLogger(out, ve, true)}

	} else if verbose < 4 {
		ve = slog.LevelInfo
	} else if verbose < 8 {
		ve = slog.LevelWarn
	} else {
		ve = slog.LevelError
	}

	return &Logger{logger: newLogger(out, ve, false)}
}

func (l *Logger) Print(msg string, args ...any) {
	args = append([]any{"package", l.moduleName}, args...)
	l.logger.Info(msg, args...)
}

func (l *Logger) Success(msg string, args ...any) {
	color.Set(color.FgGreen, color.Bold)
	defer color.Unset()
	args = append([]any{"package", l.moduleName}, args...)
	l.logger.Info(msg, args...)
}

func (l *Logger) Note(msg string, args ...any) {
	color.Set(color.FgBlue, color.Bold)
	defer color.Unset()
	args = append([]any{"package", l.moduleName}, args...)
	l.logger.Info(msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	defer color.Unset()
	args = append([]any{"package", l.moduleName}, args...)
	l.logger.Debug(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	color.Set(color.FgHiRed, color.Bold)
	defer color.Unset()
	args = append([]any{"package", l.moduleName}, args...)
	l.logger.Error(msg, args...)
}

func (l *Logger) NewError(format string, args ...any) error {
	l.Debug(format, args...)
	args = append([]any{"package", l.moduleName}, args...)
	return fmt.Errorf(format, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	color.Set(color.FgYellow, color.Bold)
	defer color.Unset()
	args = append([]any{"package", l.moduleName}, args...)
	l.logger.Warn(msg, args...)
}

func (l *Logger) Table(data []cloudController.AllClusterData) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Name", "Provider", "Nodes", "Type", "K8s")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, row := range data {
		node := ""
		if row.Type == "ha" {
			node = fmt.Sprintf("cp: %d\nwp: %d\nds: %d\nlb: 1", row.NoCP, row.NoWP, row.NoDS)
		} else {
			node = fmt.Sprintf("wp: %d", row.NoMgt)
		}
		tbl.AddRow(row.Name, string(row.Provider)+"("+row.Region+")", node, row.Type, string(row.K8sDistro))
	}

	tbl.Print()
}
