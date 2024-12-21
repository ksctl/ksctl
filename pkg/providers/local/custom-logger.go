// Copyright 2024 ksctl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package local

import (
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	klog "sigs.k8s.io/kind/pkg/log"
	"sync"
)

type customLogger struct {
	level  int32
	Logger types.LoggerFactory
	mu     sync.Mutex
}

func (l *customLogger) Enabled() bool {
	return false
}

func (l *customLogger) Info(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Logger.ExternalLogHandler(localCtx, consts.LogInfo, message)
}

func (l *customLogger) Infof(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Logger.ExternalLogHandlerf(localCtx, consts.LogInfo, format, args...)
}

func (l *customLogger) Warn(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Logger.ExternalLogHandler(localCtx, consts.LogWarning, message)
}

func (l *customLogger) Warnf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Logger.ExternalLogHandlerf(localCtx, consts.LogWarning, format, args...)
}

func (l *customLogger) Error(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Logger.ExternalLogHandler(localCtx, consts.LogError, message)
}

func (l *customLogger) Errorf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Logger.ExternalLogHandlerf(localCtx, consts.LogError, format, args...)
}

func (l *customLogger) Enable(flag bool) {}

func (l *customLogger) V(level klog.Level) klog.InfoLogger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = int32(level)
	return l
}

func (l *customLogger) WithValues(keysAndValues ...interface{}) klog.Logger {
	return l
}
