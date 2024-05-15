package local

import (
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	klog "sigs.k8s.io/kind/pkg/log"
)

type customLogger struct {
	level  int32
	Logger types.LoggerFactory
}

func (l *customLogger) Enabled() bool {
	return false
}

func (l *customLogger) Info(message string) {
	l.Logger.ExternalLogHandler(localCtx, consts.LOG_INFO, message)
}

func (l *customLogger) Infof(format string, args ...any) {
	l.Logger.ExternalLogHandlerf(localCtx, consts.LOG_INFO, format, args...)
}

func (l *customLogger) Warn(message string) {
	l.Logger.ExternalLogHandler(localCtx, consts.LOG_WARNING, message)
}

func (l *customLogger) Warnf(format string, args ...interface{}) {
	l.Logger.ExternalLogHandlerf(localCtx, consts.LOG_WARNING, format, args...)
}

func (l *customLogger) Error(message string) {
	l.Logger.ExternalLogHandler(localCtx, consts.LOG_ERROR, message)
}

func (l *customLogger) Errorf(format string, args ...interface{}) {
	l.Logger.ExternalLogHandlerf(localCtx, consts.LOG_ERROR, format, args...)
}

func (l *customLogger) Enable(flag bool) {}

func (l *customLogger) V(level klog.Level) klog.InfoLogger {
	l.level = int32(level)
	return l
}

func (l *customLogger) WithValues(keysAndValues ...interface{}) klog.Logger {
	return l
}
