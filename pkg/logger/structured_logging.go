// Copyright 2024 Ksctl Authors
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

package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"strings"

	"github.com/fatih/color"
	"github.com/ksctl/ksctl/v2/pkg/consts"
)

type StructuredLog struct {
	logger *slog.Logger
}

const (
	limitCol = 80
)

func getPackageName(ctx context.Context) string {
	if v, ok := ctx.Value(consts.KsctlModuleNameKey).(string); ok {
		return v
	} else {
		return "!!NOT_SET"
	}
}

func (l *StructuredLog) ExternalLogHandler(ctx context.Context, msgType CustomExternalLogLevel, message string) {
	_m := ""
	switch msgType {
	case LogDebug:
		_m = "DEBUG"
	case LogError:
		_m = "ERRO"
	case LogSuccess:
		_m = "SUCC"
	case LogWarning:
		_m = "WARN"
	default:
		_m = "INFO"
	}
	l.logger.Info(message, "component", getPackageName(ctx), "msgType", _m)
}

func (l *StructuredLog) ExternalLogHandlerf(ctx context.Context, msgType CustomExternalLogLevel, format string, args ...interface{}) {
	_m := ""
	switch msgType {
	case LogDebug:
		_m = "DEBUG"
	case LogError:
		_m = "ERR"
	case LogSuccess:
		_m = "SUCC"
	case LogWarning:
		_m = "WARN"
	default:
		_m = "INFO"
	}
	l.logger.Info(fmt.Sprintf(format, args...), "component", getPackageName(ctx), "msgType", _m)
}

func newLogger(out io.Writer, ver slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: ver,
	}))
}

func NewStructuredLogger(verbose int, out io.Writer) *StructuredLog {
	// LevelDebug Level = -4
	// LevelInfo  Level = 0
	// LevelWarn  Level = 4
	// LevelError Level = 8

	var ve slog.Level

	if verbose < 0 {
		ve = slog.LevelDebug
	} else if verbose < 4 {
		ve = slog.LevelInfo
	} else if verbose < 8 {
		ve = slog.LevelWarn
	} else {
		ve = slog.LevelError
	}

	return &StructuredLog{logger: newLogger(out, ve)}
}

func formGroups(v ...any) (format string, vals []any) {
	if len(v) == 0 {
		return "", nil
	}
	_format := strings.Builder{}

	defer func() {
		format = strings.TrimSpace(_format.String())
	}()
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
	return
}

func (l *StructuredLog) logErrorf(msg string, args ...any) error {
	format, _args := formGroups(args...)

	var errMsg error
	if _args == nil {
		errMsg = fmt.Errorf(msg + " " + format)
	} else {
		errMsg = fmt.Errorf(msg+" "+format, _args...)
	}

	return errMsg
}

func (l *StructuredLog) Print(ctx context.Context, msg string, args ...any) {
	args = append([]any{"component", getPackageName(ctx)}, args...)
	l.logger.Info(msg, args...)
}

func (l *StructuredLog) Success(ctx context.Context, msg string, args ...any) {
	color.Set(color.FgGreen, color.Bold)
	defer color.Unset()
	args = append([]any{"component", getPackageName(ctx), "msgType", "SUCCESS"}, args...)
	l.logger.Info(msg, args...)
}

func (l *StructuredLog) Note(ctx context.Context, msg string, args ...any) {
	color.Set(color.FgBlue, color.Bold)
	defer color.Unset()
	args = append([]any{"component", getPackageName(ctx), "msgType", "NOTE"}, args...)
	l.logger.Info(msg, args...)
}

func (l *StructuredLog) Debug(ctx context.Context, msg string, args ...any) {
	defer color.Unset()
	args = append([]any{"component", getPackageName(ctx)}, args...)
	l.logger.Debug(msg, args...)
}

func (l *StructuredLog) Error(msg string, args ...any) {
	color.Set(color.FgHiRed, color.Bold)
	defer color.Unset()

	l.logger.Error(msg, args...)
}

func (l *StructuredLog) NewError(ctx context.Context, format string, args ...any) error {
	args = append([]any{"component", getPackageName(ctx)}, args...)
	return l.logErrorf(format, args...)
}

func (l *StructuredLog) Warn(ctx context.Context, msg string, args ...any) {
	color.Set(color.FgYellow, color.Bold)
	defer color.Unset()
	args = append([]any{"component", getPackageName(ctx), "msgType", "WARN"}, args...)
	l.logger.Warn(msg, args...)
}

func (l *StructuredLog) Table(ctx context.Context, headers []string, data [][]string) {
	l.Success(ctx, "table content", "headers", headers, "data", data)
}

func (l *StructuredLog) Box(ctx context.Context, title string, lines string) {

	l.Print(ctx, title, "details", addLineTerminationForLongStrings(lines))
}
