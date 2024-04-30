package logger

import (
	"fmt"
	"time"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

func (l *Logger) ExternalLogHandler(msgType consts.CustomExternalLogLevel, message string) {
	fmt.Printf("%s (package: %s) [%s] %v\n", time.Now(), l.moduleName, msgType, message)
}

func (l *Logger) ExternalLogHandlerf(msgType consts.CustomExternalLogLevel, format string, args ...interface{}) {
	prefix := fmt.Sprintf("%s (package: %s) [%s] ", time.Now(), l.moduleName, msgType)
	format = prefix + format
	fmt.Printf(format, args...)
}
