package resources

import (
	"io"
)

type LoggerFactory interface {
	Info(LoggerMsgType, ...any)
	Debug(LoggerMsgType, ...any)
	Infof(LoggerMsgType, string, ...any)
	Debugf(LoggerMsgType, string, ...any)

	New(int, io.Writer) error
	AppendPrefix(string)
	ResetPrefix()
	PopPrefix()
}

type LoggerMsgType string

const (
	MsgTypeSuccess = LoggerMsgType("success")
	MsgTypeWarn    = LoggerMsgType("warn")
	MsgTypeInfo    = LoggerMsgType("info")
	MsgTypePrint   = LoggerMsgType("print")
	MsgTypeError   = LoggerMsgType("error")
)
