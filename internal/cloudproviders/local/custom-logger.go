package local

import (
	"github.com/kubesimplify/ksctl/pkg/resources"
	klog "sigs.k8s.io/kind/pkg/log"
)

type CustomLogger struct {
	Logger resources.LoggerFactory
}

func (l CustomLogger) Enabled() bool {
	return false
}

func (l CustomLogger) Info(message string) {
	l.Logger.Print(message)
}

func (l CustomLogger) Infof(format string, args ...any) {
	l.Logger.Print(format, args...)
}

func (l CustomLogger) Warn(message string) {
	l.Logger.Warn(message)
}

func (l CustomLogger) Warnf(format string, args ...interface{}) {
	l.Logger.Warn(format, args...)
}

func (l CustomLogger) Error(message string) {
	l.Logger.Error(message)
}

func (l CustomLogger) Errorf(format string, args ...interface{}) {
	l.Logger.Error(format, args...)
}

func (l CustomLogger) Enable(flag bool) {}

func (l CustomLogger) V(level klog.Level) klog.InfoLogger {
	return l
}

func (l CustomLogger) WithValues(keysAndValues ...interface{}) klog.Logger {
	return l
}
