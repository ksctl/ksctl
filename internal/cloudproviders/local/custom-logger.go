package local

import (
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
	klog "sigs.k8s.io/kind/pkg/log"
)

type CustomLogger struct {
	level  int32
	Logger resources.LoggerFactory
}

func (l *CustomLogger) Enabled() bool {
	return false
}

func (l *CustomLogger) Info(message string) {
	l.Logger.ExternalLogHandler(consts.LOG_INFO, message)
}

func (l *CustomLogger) Infof(format string, args ...any) {
	l.Logger.ExternalLogHandlerf(consts.LOG_INFO, format, args...)
}

func (l *CustomLogger) Warn(message string) {
	l.Logger.ExternalLogHandler(consts.LOG_WARNING, message)
}

func (l *CustomLogger) Warnf(format string, args ...interface{}) {
	l.Logger.ExternalLogHandlerf(consts.LOG_WARNING, format, args...)
}

func (l *CustomLogger) Error(message string) {
	l.Logger.ExternalLogHandler(consts.LOG_ERROR, message)
}

func (l *CustomLogger) Errorf(format string, args ...interface{}) {
	l.Logger.ExternalLogHandlerf(consts.LOG_ERROR, format, args...)
}

func (l *CustomLogger) Enable(flag bool) {}

func (l *CustomLogger) V(level klog.Level) klog.InfoLogger {
	l.level = int32(level)
	return l
}

func (l *CustomLogger) WithValues(keysAndValues ...interface{}) klog.Logger {
	return l
}
