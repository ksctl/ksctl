package logger

import (
	"context"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
	"sync"

	box "github.com/Delta456/box-cli-maker/v2"
	"github.com/fatih/color"
	"github.com/ksctl/ksctl/pkg/types"
	cloudController "github.com/ksctl/ksctl/pkg/types/controllers/cloud"
	"github.com/rodaine/table"

	"time"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

type GeneralLog struct {
	mu      *sync.Mutex
	writter io.Writer
	level   uint
}

func (l *GeneralLog) ExternalLogHandler(ctx context.Context, msgType consts.CustomExternalLogLevel, message string) {
	l.log(false, false, ctx, msgType, message)
}

func (l *GeneralLog) ExternalLogHandlerf(ctx context.Context, msgType consts.CustomExternalLogLevel, format string, args ...interface{}) {
	l.log(false, false, ctx, msgType, format, args...)
}

func formGroups(disableContext bool, ctx context.Context, v ...any) (format string, vals []any) {
	if len(v) == 0 {
		return "\n", nil
	}
	_format := strings.Builder{}

	defer func() {
		format = _format.String()
	}()
	if !disableContext {
		_format.WriteString("component=%s ")
		vals = append(vals, color.MagentaString(getPackageName(ctx)))
	}
	i := 0
	for ; i+1 < len(v); i += 2 {
		if !reflect.TypeOf(v[i+1]).Implements(reflect.TypeOf((*error)(nil)).Elem()) &&
			(reflect.TypeOf(v[i+1]).Kind() == reflect.Interface ||
				reflect.TypeOf(v[i+1]).Kind() == reflect.Ptr ||
				reflect.TypeOf(v[i+1]).Kind() == reflect.Struct) {
			_format.WriteString(fmt.Sprintf("%s", v[i]) + "=%#v ")
		} else {
			_format.WriteString(fmt.Sprintf("%s", v[i]) + "=%v ")
		}

		vals = append(vals, v[i+1])
	}

	for ; i < len(v); i++ {
		_format.WriteString("!!EXTRA:%v ")
		vals = append(vals, v[i])
	}
	_format.WriteString("\n")
	return
}

func isLogEnabled(level uint, msgType consts.CustomExternalLogLevel) bool {
	if msgType == consts.LOG_DEBUG {
		return level >= 9
	}
	return true
}

func (l *GeneralLog) logErrorf(disableContext bool, disablePrefix bool, ctx context.Context, msg string, args ...any) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !disablePrefix {
		prefix := fmt.Sprintf("%s[%s] ", getTime(l.level), consts.LOG_ERROR)
		msg = prefix + msg
	}
	format, _args := formGroups(disableContext, ctx, args...)

	var errMsg error
	if _args == nil {
		errMsg = fmt.Errorf(msg + " " + format)
	} else {
		errMsg = fmt.Errorf(msg+" "+format, _args...)
	}

	return errMsg
}

func (l *GeneralLog) log(disableContext bool, useGroupFormer bool, ctx context.Context, msgType consts.CustomExternalLogLevel, msg string, args ...any) {
	if !isLogEnabled(l.level, msgType) {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	prefix := fmt.Sprintf("%s[%s] ", getTime(l.level), msgType)

	if useGroupFormer {
		msg = prefix + msg
		format, _args := formGroups(disableContext, ctx, args...)
		if _args == nil {
			fmt.Fprint(l.writter, msg+" "+format)
		} else {
			fmt.Fprintf(l.writter, msg+" "+format, _args...)
		}
	} else {
		args = append([]any{color.MagentaString(getPackageName(ctx))}, args...)
		fmt.Fprintf(l.writter, prefix+"component=%s "+msg+"\n", args...)
	}
}

func getTime(level uint) string {
	if level < 9 {
		return ""
	}
	t := time.Now()
	return fmt.Sprintf("%d:%d:%d ", t.Hour(), t.Minute(), t.Second())
}

func NewGeneralLogger(verbose int, out io.Writer) types.LoggerFactory {

	var ve uint

	if verbose < 0 {
		ve = 9
	}

	return &GeneralLog{
		writter: out,
		level:   ve,
		mu:      new(sync.Mutex),
	}
}

func (l *GeneralLog) Print(ctx context.Context, msg string, args ...any) {
	l.log(false, true, ctx, consts.LOG_INFO, msg, args...)
}

func (l *GeneralLog) Success(ctx context.Context, msg string, args ...any) {
	l.log(false, true, ctx, consts.LOG_SUCCESS, msg, args...)
}

func (l *GeneralLog) Note(ctx context.Context, msg string, args ...any) {
	l.log(false, true, ctx, consts.LOG_NOTE, msg, args...)
}

func (l *GeneralLog) Debug(ctx context.Context, msg string, args ...any) {
	l.log(false, true, ctx, consts.LOG_DEBUG, msg, args...)
}

// TODO: Depricate the context for Error() method
func (l *GeneralLog) Error(ctx context.Context, msg string, args ...any) {
	l.log(true, true, ctx, consts.LOG_ERROR, msg, args...)
}

func (l *GeneralLog) NewError(ctx context.Context, msg string, args ...any) error {
	return l.logErrorf(false, true, ctx, msg, args...)
}

func (l *GeneralLog) Warn(ctx context.Context, msg string, args ...any) {
	l.log(false, true, ctx, consts.LOG_WARNING, msg, args...)
}

func (l *GeneralLog) Table(ctx context.Context, data []cloudController.AllClusterData) {
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

func (l *GeneralLog) Box(ctx context.Context, title string, lines string) {
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
