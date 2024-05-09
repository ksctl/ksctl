package logger

import (
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
	"sync"

	box "github.com/Delta456/box-cli-maker/v2"
	"github.com/fatih/color"
	"github.com/ksctl/ksctl/pkg/resources"
	cloudController "github.com/ksctl/ksctl/pkg/resources/controllers/cloud"
	"github.com/rodaine/table"

	"time"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

type GeneralLog struct {
	mu         *sync.Mutex
	writter    io.Writer
	moduleName string
	level      uint
}

func (l *GeneralLog) ExternalLogHandler(msgType consts.CustomExternalLogLevel, message string) {
	l.log(msgType, message)
}

func (l *GeneralLog) ExternalLogHandlerf(msgType consts.CustomExternalLogLevel, format string, args ...interface{}) {
	l.log(msgType, format, args...)
}

func formGroups(l *GeneralLog, v ...any) (format string, vals []any) {
	if len(v) == 0 {
		return "\n", nil
	}
	_format := strings.Builder{}

	defer func() {
		format = _format.String()
	}()
	_format.WriteString("component=%s ")
	vals = append(vals, color.MagentaString(l.moduleName))
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

func (l *GeneralLog) log(msgType consts.CustomExternalLogLevel, msg string, args ...any) {
	if !isLogEnabled(l.level, msgType) {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	prefix := fmt.Sprintf("%s[%s] ", getTime(l.level), msgType)
	msg = prefix + msg
	format, _args := formGroups(l, args...)
	if _args == nil {
		fmt.Fprint(l.writter, msg+" "+format)
	} else {
		fmt.Fprintf(l.writter, msg+" "+format, _args...)
	}
}

func (l *GeneralLog) SetPackageName(m string) {
	l.moduleName = m
}

func getTime(level uint) string {
	if level < 9 {
		return ""
	}
	t := time.Now()
	return fmt.Sprintf("%d:%d:%d ", t.Hour(), t.Minute(), t.Second())
}

func NewGeneralLogger(verbose int, out io.Writer) resources.LoggerFactory {

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

func (l *GeneralLog) Print(msg string, args ...any) {
	l.log(consts.LOG_INFO, msg, args...)
}

func (l *GeneralLog) Success(msg string, args ...any) {
	l.log(consts.LOG_SUCCESS, msg, args...)
}

func (l *GeneralLog) Note(msg string, args ...any) {
	l.log(consts.LOG_NOTE, msg, args...)
}

func (l *GeneralLog) Debug(msg string, args ...any) {
	l.log(consts.LOG_DEBUG, msg, args...)
}

func (l *GeneralLog) Error(msg string, args ...any) {
	l.log(consts.LOG_ERROR, msg, args...)
}

func (l *GeneralLog) NewError(format string, args ...any) error {
	// TODO: add the package info!
	return fmt.Errorf(format, args...)
}

func (l *GeneralLog) Warn(msg string, args ...any) {
	l.log(consts.LOG_WARNING, msg, args...)
}

func (l *GeneralLog) Table(data []cloudController.AllClusterData) {
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

func (l *GeneralLog) Box(title string, lines string) {
	px := 4

	if len(title) >= 2*px+len(lines) {
		// some maths
		px = int(math.Ceil(float64(len(title)-len(lines))/2)) + 1
	}

	l.Debug("PostUpdate Box", "px", px, "title", len(title), "lines", len(lines))

	Box := box.New(box.Config{
		Px:       px,
		Py:       2,
		Type:     "Bold",
		TitlePos: "Top",
		Color:    "Cyan"})

	Box.Println(title, addLineTerminationForLongStrings(lines))
}
