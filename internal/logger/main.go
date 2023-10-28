package logger

import (
	"errors"
	"github.com/fatih/color"
	"github.com/kubesimplify/ksctl/pkg/resources"
	"io"
	"log"
	"strings"
)

type prefixes []string

type Logger struct {
	userVerboseLevel int
	prefixHistory    prefixes
	customLog        *log.Logger
}

func (p prefixes) String() string {
	var str strings.Builder

	for _, s := range p {
		str.WriteString("[" + s + "]")
	}
	return str.String()
}

func (l *Logger) New(verbose int, out io.Writer) error {
	if verbose > 0 && verbose < 9 {
		l.userVerboseLevel = verbose
		l.customLog = log.New(out, "[ksctl]", verbose)
		return nil
	} else {
		return errors.New("invalid verbosity level")
	}
}

func (l *Logger) AppendPrefix(pre string) {
	l.prefixHistory = append(l.prefixHistory, pre)
	l.customLog.SetPrefix(l.prefixHistory.String())
}

func (l *Logger) PopPrefix() {
	l.prefixHistory = l.prefixHistory[:len(l.prefixHistory)-1]
	l.customLog.SetPrefix(l.prefixHistory.String())
}

func (l *Logger) ResetPrefix() {
	l.prefixHistory = prefixes{}
	l.prefixHistory = append(l.prefixHistory, "ksctl")
	l.customLog.SetPrefix(l.prefixHistory.String())
}

func setMessagePrefix(msg resources.LoggerMsgType, format *string, args *[]any) {
	msgCode := ""
	switch msg {
	case resources.MsgTypeSuccess:
		color.Set(color.FgGreen, color.Bold)
		msgCode = "[SUCCESS]"
	case resources.MsgTypeError:
		color.Set(color.FgHiRed, color.Bold)
		msgCode = "[ERR]"
	case resources.MsgTypeWarn:
		color.Set(color.FgYellow, color.Bold)
		msgCode = "[WARN]"
	}
	if format != nil {
		*format = msgCode + " " + strings.TrimSpace(*format)
	} else {
		*args = append([]any{msgCode}, *args...)
	}
}

func (l *Logger) Info(msg resources.LoggerMsgType, args ...any) {
	defer color.Unset()

	if l.userVerboseLevel > 3 {
		// skip
		return
	}
	setMessagePrefix(msg, nil, &args)
	// create a seeprate function which will append the ...any with the string
	l.customLog.Println(args...)
}

func (l *Logger) Debug(msg resources.LoggerMsgType, args ...any) {
	defer color.Unset()

	if l.userVerboseLevel <= 3 || l.userVerboseLevel > 9 {
		// skip
		return
	}
	setMessagePrefix(msg, nil, &args)
	l.customLog.Println(args...)
}

func (l *Logger) Infof(msg resources.LoggerMsgType, format string, args ...any) {
	defer color.Unset()

	if l.userVerboseLevel > 3 {
		// skip
		return
	}
	setMessagePrefix(msg, &format, &args)
	l.customLog.Printf(format, args...)
}

func (l *Logger) Debugf(msg resources.LoggerMsgType, format string, args ...any) {
	defer color.Unset()

	if l.userVerboseLevel <= 3 || l.userVerboseLevel > 9 {
		// skip
		return
	}
	setMessagePrefix(msg, &format, &args)
	l.customLog.Printf(format, args...)
}
