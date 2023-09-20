package local

import (
	"fmt"

	"github.com/kubesimplify/ksctl/api/resources"
	"sigs.k8s.io/kind/pkg/log"
)

type CustomLogger struct {
	StorageDriver resources.StorageFactory
}

func (l CustomLogger) Enabled() bool {
	return true
}

func (l CustomLogger) Info(message string) {
	l.StorageDriver.Logger().Print("[local]", message)
}

func (l CustomLogger) Infof(format string, args ...interface{}) {
	l.StorageDriver.Logger().Print(fmt.Sprintf("[Local]: "+format+"\n", args...))
}

func (l CustomLogger) Warn(message string) {
	l.StorageDriver.Logger().Warn("[local]: ", message)
}

func (l CustomLogger) Warnf(format string, args ...interface{}) {
	l.StorageDriver.Logger().Warn(fmt.Sprintf("[local]: "+format+"\n", args...))
}

func (l CustomLogger) Error(message string) {
	l.StorageDriver.Logger().Err("[local]", message)
}

func (l CustomLogger) Errorf(format string, args ...interface{}) {
	l.StorageDriver.Logger().Warn(fmt.Sprintf("[local]: "+format+"\n", args...))
}

func (l CustomLogger) Enable(flag bool) {}

func (l CustomLogger) V(level log.Level) log.InfoLogger {
	return l
}

func (l CustomLogger) WithValues(keysAndValues ...interface{}) log.Logger {
	return l
}
