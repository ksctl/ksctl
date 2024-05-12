package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"

	box "github.com/Delta456/box-cli-maker/v2"
	"github.com/fatih/color"
	"github.com/ksctl/ksctl/pkg/types"
	cloudController "github.com/ksctl/ksctl/pkg/types/controllers/cloud"
	"github.com/rodaine/table"

	"time"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

type StructuredLog struct {
	logger *slog.Logger
}

const (
	LimitCol = 80
)

// FIXME: Missing a mutex

func (l *StructuredLog) ExternalLogHandler(ctx context.Context, msgType consts.CustomExternalLogLevel, message string) {
	fmt.Printf("%s (package: %s) [%s] %v\n", time.Now().Format(time.ANSIC), ctx.Value(consts.ContextModuleNameKey), msgType, message)
}

func (l *StructuredLog) ExternalLogHandlerf(ctx context.Context, msgType consts.CustomExternalLogLevel, format string, args ...interface{}) {
	prefix := fmt.Sprintf("%s (package: %s) [%s] ", time.Now().Format(time.ANSIC), ctx.Value(consts.ContextModuleNameKey), msgType)
	format = prefix + format
	fmt.Printf(format, args...)
}

func newLogger(out io.Writer, ver slog.Level, _ bool) *slog.Logger {
	// if debug {
	return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: ver,
	}))
	// }
	// return slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{
	// 	Level: ver,
	// }))
}

func NewStructuredLogger(verbose int, out io.Writer) types.LoggerFactory {
	// LevelDebug Level = -4
	// LevelInfo  Level = 0
	// LevelWarn  Level = 4
	// LevelError Level = 8

	var ve slog.Level

	if verbose < 0 {
		ve = slog.LevelDebug

		return &StructuredLog{logger: newLogger(out, ve, true)}

	} else if verbose < 4 {
		ve = slog.LevelInfo
	} else if verbose < 8 {
		ve = slog.LevelWarn
	} else {
		ve = slog.LevelError
	}

	return &StructuredLog{logger: newLogger(out, ve, false)}
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
	args = append([]any{"package", ctx.Value(consts.ContextModuleNameKey), "msgType", "NOTE"}, args...)
	l.logger.Info(msg, args...)
}

func (l *StructuredLog) Debug(ctx context.Context, msg string, args ...any) {
	defer color.Unset()
	args = append([]any{"package", ctx.Value(consts.ContextModuleNameKey)}, args...)
	l.logger.Debug(msg, args...)
}

func (l *StructuredLog) Error(ctx context.Context, msg string, args ...any) {
	color.Set(color.FgHiRed, color.Bold)
	defer color.Unset()
	args = append([]any{"package", ctx.Value(consts.ContextModuleNameKey)}, args...)

	msgStrErr := fmt.Sprintf("%v", args)
	l.logger.Error(msg, "Reason", msgStrErr)
}

func (l *StructuredLog) NewError(ctx context.Context, format string, args ...any) error {
	args = append([]any{"err_package", ctx.Value(consts.ContextModuleNameKey)}, args...)
	return fmt.Errorf(format, args...)
}

func (l *StructuredLog) Warn(ctx context.Context, msg string, args ...any) {
	color.Set(color.FgYellow, color.Bold)
	defer color.Unset()
	args = append([]any{"package", ctx.Value(consts.ContextModuleNameKey), "msgType", "WARN"}, args...)
	l.logger.Warn(msg, args...)
}

func (l *StructuredLog) Table(ctx context.Context, data []cloudController.AllClusterData) {
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

func (l *StructuredLog) Box(ctx context.Context, title string, lines string) {
	px := 4

	if len(title) >= 2*px+len(lines) {
		// some maths
		px = int(math.Ceil(float64(len(title)-len(lines))/2)) + 1
	}

	l.Debug(ctx, "PostUpdate Box", "px", px, "title", len(title), "lines", len(lines))

	Box := box.New(box.Config{
		Px:       px,
		Py:       2,
		Type:     "Bold",
		TitlePos: "Top",
		Color:    "Cyan"})

	Box.Println(title, addLineTerminationForLongStrings(lines))
}
