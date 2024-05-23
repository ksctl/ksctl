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
	l.Logger.ExternalLogHandler(localCtx, consts.LogInfo, message)
}

func (l *customLogger) Infof(format string, args ...any) {
	l.Logger.ExternalLogHandlerf(localCtx, consts.LogInfo, format, args...)
}

func (l *customLogger) Warn(message string) {
	l.Logger.ExternalLogHandler(localCtx, consts.LogWarning, message)
}

func (l *customLogger) Warnf(format string, args ...interface{}) {
	l.Logger.ExternalLogHandlerf(localCtx, consts.LogWarning, format, args...)
}

func (l *customLogger) Error(message string) {
	l.Logger.ExternalLogHandler(localCtx, consts.LogError, message)
}

func (l *customLogger) Errorf(format string, args ...interface{}) {
	l.Logger.ExternalLogHandlerf(localCtx, consts.LogError, format, args...)
}

func (l *customLogger) Enable(flag bool) {}

func (l *customLogger) V(level klog.Level) klog.InfoLogger {
	l.level = int32(level)
	return l
}

func (l *customLogger) WithValues(keysAndValues ...interface{}) klog.Logger {
	return l
}
