// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	// DefaultLogger is the default logger
	DefaultLogger Logger

	// DefaultRegistry is the default logging registry
	DefaultRegistry Registry
)

var (
	Trace  = logLvlFn(TraceLevel)
	Tracef = logfLvlFn(TraceLevel)
	Debug  = logLvlFn(DebugLevel)
	Debugf = logfLvlFn(DebugLevel)
	Info   = logLvlFn(InfoLevel)
	Infof  = logfLvlFn(InfoLevel)
	Warn   = logLvlFn(WarnLevel)
	Warnf  = logfLvlFn(WarnLevel)
	Error  = logLvlFn(ErrorLevel)
	Errorf = logfLvlFn(ErrorLevel)
	Fatal  = log.Fatal
	Fatalf = log.Fatalf
	Panic  = log.Panic
	Panicf = log.Panicf
)

func logLvlFn(lvl LogLevel) func(...interface{}) {
	return func(args ...interface{}) {
		log.Printf("%s: %s", strings.ToUpper(lvl.String()), fmt.Sprint(args...))
	}
}

func logfLvlFn(lvl LogLevel) func(string, ...interface{}) {
	return func(format string, args ...interface{}) {
		log.Printf("%s: %s", strings.ToUpper(lvl.String()), fmt.Sprintf(format, args...))
	}
}

// LogWithLevel allows to log with different log levels
type LogWithLevel interface {
	WithField(key string, value interface{}) LogWithLevel
	WithFields(fields Fields) LogWithLevel
	WithError(err error) LogWithLevel

	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	Printf(format string, args ...interface{})

	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
	Print(args ...interface{})

	Traceln(args ...interface{})
	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Println(args ...interface{})
	Warnln(args ...interface{})
	Warningln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
	Panicln(args ...interface{})
}

// Logger provides logging capabilities
type Logger interface {
	LogWithLevel

	// GetName returns the logger name
	GetName() string
	// SetLevel modifies the log level
	SetLevel(level LogLevel)
	// GetLevel returns currently set log level
	GetLevel() LogLevel
	// AddHook adds hook to logger
	AddHook(hook logrus.Hook)
	// SetOutput sets output writer
	SetOutput(out io.Writer)
	// SetFormatter sets custom formatter
	SetFormatter(formatter logrus.Formatter)
}

// LoggerFactory is API for the plugins that want to create their own loggers.
type LoggerFactory interface {
	NewLogger(name string) Logger
}

// Registry groups multiple Logger instances and allows to mange their log levels.
type Registry interface {
	LoggerFactory

	// ListLoggers returns a map (loggerName => log level)
	ListLoggers() map[string]string
	// SetLevel modifies log level of selected logger in the registry
	SetLevel(logger, level string) error
	// GetLevel returns the currently set log level of the logger from registry
	GetLevel(logger string) (string, error)
	// Lookup returns a logger instance identified by name from registry
	Lookup(loggerName string) (logger Logger, found bool)
	// ClearRegistry removes all loggers except the default one from registry
	ClearRegistry()
	// AddHook stores hooks from log manager to be used for new loggers
	AddHook(hook logrus.Hook)
}

// Fields is a type accepted by WithFields method.
type Fields map[string]interface{}

// LogLevel defines severity of log entry.
type LogLevel uint32

const (
	// PanicLevel - highest level of severity. Logs and then calls panic with the message passed in.
	PanicLevel LogLevel = iota
	// FatalLevel - logs and then calls `os.Exit(1)`.
	FatalLevel
	// ErrorLevel - used for errors that should definitely be noted.
	ErrorLevel
	// WarnLevel - non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel - general operational entries about what's going on inside the application.
	InfoLevel
	// DebugLevel - enabled for debugging, very verbose logging.
	DebugLevel
	// TraceLevel - extra level for debugging specific parts.
	TraceLevel
)

// Convert the LogLevel to a string.
func (level LogLevel) String() string {
	if b, err := level.MarshalText(); err == nil {
		return string(b)
	}
	return fmt.Sprintf("UnknownLevel(%d)", level)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (level *LogLevel) UnmarshalText(text []byte) error {
	l, err := ParseLogLevel(string(text))
	if err != nil {
		return err
	}
	*level = LogLevel(l)
	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (level LogLevel) MarshalText() ([]byte, error) {
	switch level {
	case TraceLevel:
		return []byte("trace"), nil
	case DebugLevel:
		return []byte("debug"), nil
	case InfoLevel:
		return []byte("info"), nil
	case WarnLevel:
		return []byte("warn"), nil
	case ErrorLevel:
		return []byte("error"), nil
	case FatalLevel:
		return []byte("fatal"), nil
	case PanicLevel:
		return []byte("panic"), nil
	}
	return nil, fmt.Errorf("invalid log level %d", level)
}

// ParseLogLevel parses the string as log level.
func ParseLogLevel(level string) (LogLevel, error) {
	switch strings.ToLower(level) {
	case "panic":
		return PanicLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	case "trace":
		return TraceLevel, nil
	}
	return InfoLevel, fmt.Errorf("invalid log level: %q", level)
}
