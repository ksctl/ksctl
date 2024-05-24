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
		_format.WriteString(color.HiBlackString("component=") + "%s ")
		vals = append(vals, getPackageName(ctx))
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
	if msgType == consts.LogDebug {
		return level >= 9
	}
	return true
}

func (l *GeneralLog) logErrorf(disableContext bool, disablePrefix bool, ctx context.Context, msg string, args ...any) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !disablePrefix {
		prefix := fmt.Sprintf("%s%s ", getTime(l.level), consts.LogError)
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
	prefix := fmt.Sprintf("%s%s ", getTime(l.level), msgType)

	if useGroupFormer {
		msg = prefix + msg
		format, _args := formGroups(disableContext, ctx, args...)
		if _args == nil {
			if disableContext && msgType == consts.LogError {
				l.boxBox(
					"Error", msg+" "+format, "Red")
				return
			}
			fmt.Fprint(l.writter, msg+" "+format)
		} else {
			if disableContext && msgType == consts.LogError {
				l.boxBox(
					"Error", fmt.Sprintf(msg+" "+format, _args...), "Red")
				return
			}
			fmt.Fprintf(l.writter, msg+" "+format, _args...)
		}
	} else {
		args = append([]any{getPackageName(ctx)}, args...)
		fmt.Fprintf(l.writter, prefix+color.HiBlackString("component=")+"%s "+msg+"\n", args...)
	}
}

func getTime(level uint) string {
	// if level < 9 {
	// 	return ""
	// }
	t := time.Now()
	return color.HiBlackString(fmt.Sprintf("%d:%d:%d ", t.Hour(), t.Minute(), t.Second()))
}

func NewGeneralLogger(verbose int, out io.Writer) *GeneralLog {

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
	l.log(false, true, ctx, consts.LogInfo, msg, args...)
}

func (l *GeneralLog) Success(ctx context.Context, msg string, args ...any) {
	l.log(false, true, ctx, consts.LogSuccess, msg, args...)
}

func (l *GeneralLog) Note(ctx context.Context, msg string, args ...any) {
	l.log(false, true, ctx, consts.LogNote, msg, args...)
}

func (l *GeneralLog) Debug(ctx context.Context, msg string, args ...any) {
	l.log(false, true, ctx, consts.LogDebug, msg, args...)
}

func (l *GeneralLog) Error(ctx context.Context, msg string, args ...any) {
	l.log(true, true, ctx, consts.LogError, msg, args...)
}

func (l *GeneralLog) NewError(ctx context.Context, msg string, args ...any) error {
	return l.logErrorf(false, true, ctx, msg, args...)
}

func (l *GeneralLog) Warn(ctx context.Context, msg string, args ...any) {
	l.log(false, true, ctx, consts.LogWarning, msg, args...)
}

func (l *GeneralLog) Table(ctx context.Context, data []cloudController.AllClusterData) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("ClusterName", "Region", "ClusterType", "CloudProvider", "BootStrap", "Mgt Nodes", "Self-Mgt Nodes")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	extractVMSizes := func(v []cloudController.VMData) string {
		sizes := []string{}
		for _, _v := range v {
			sizes = append(sizes, _v.VMSize)
		}
		return strings.Join(sizes, ",")
	}

	for _, row := range data {
		mgtN, selfMgtN := "null", "null"
		if row.ClusterType == "ha" {
			selfMgtN = fmt.Sprintf(
				"cp: %d{%s}\nwp: %d{%s}\nds: %d{%s}\nlb: 1{%s}\n",
				row.NoCP,
				extractVMSizes(row.CP),
				row.NoWP,
				extractVMSizes(row.WP),
				row.NoDS,
				extractVMSizes(row.DS),
				row.LB.VMSize,
			)
		} else {
			mgtN = fmt.Sprintf("wp: %d{%s}", row.NoMgt, row.Mgt.VMSize)
		}
		tbl.AddRow(row.Name,
			row.Region,
			string(row.ClusterType),
			string(row.CloudProvider),
			string(row.K8sDistro),
			mgtN,
			selfMgtN,
		)
	}

	tbl.Print()
}

func (l *GeneralLog) boxBox(title, lines string, color string) {

	px := 4

	if len(title) >= 2*px+len(lines) {
		// some maths
		px = int(math.Ceil(float64(len(title)-len(lines))/2)) + 1
	}

	Box := box.New(box.Config{
		Px:       px,
		Py:       2,
		Type:     "Round",
		TitlePos: "Top",
		Color:    color})

	Box.Println(title, addLineTerminationForLongStrings(lines))
}

func (l *GeneralLog) Box(ctx context.Context, title string, lines string) {

	l.Debug(ctx, "PostUpdate Box", "title", len(title), "lines", len(lines))

	l.boxBox(title, lines, "Yellow")
}
