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
	"strings"

	"github.com/sirupsen/logrus"
)

// LogWithLevel allows to log with different log levels
type LogWithLevel interface {
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	Printf(format string, args ...interface{})

	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
	Print(args ...interface{})

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
	SetLevel(level Level)
	// GetLevel returns currently set log level
	GetLevel() Level
	// WithField creates one structured field
	WithField(key string, value interface{}) LogWithLevel
	// WithFields creates multiple structured fields
	WithFields(fields Fields) LogWithLevel
	// Add hook to send log to external address
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

	// List Loggers returns a map (loggerName => log level)
	ListLoggers() map[string]string
	// SetLevel modifies log level of selected logger in the registry
	SetLevel(logger, level string) error
	// GetLevel returns the currently set log level of the logger from registry
	GetLevel(logger string) (string, error)
	// Lookup returns a logger instance identified by name from registry
	Lookup(loggerName string) (logger Logger, found bool)
	// ClearRegistry removes all loggers except the default one from registry
	ClearRegistry()
	// HookConfigs stores hooks from log manager to be used for new loggers
	AddHook(hook logrus.Hook)
}

// Fields is a type accepted by WithFields method.
type Fields map[string]interface{}

// Level defines severity of log entry.
type Level uint32

const (
	// PanicLevel - highest level of severity. Logs and then calls panic with the message passed in.
	PanicLevel Level = iota
	// FatalLevel - logs and then calls `os.Exit(1)`.
	FatalLevel
	// ErrorLevel - used for errors that should definitely be noted.
	ErrorLevel
	// WarningLevel - non-critical entries that deserve eyes.
	WarningLevel
	// InfoLevel - general operational entries about what's going on inside the application.
	InfoLevel
	// DebugLevel - enabled for debugging, very verbose logging.
	DebugLevel
	// TraceLevel - extra level for debugging specific parts.
	TraceLevel
)

// Convert the Level to a string.
func (level Level) String() string {
	if b, err := level.MarshalText(); err == nil {
		return string(b)
	}
	return fmt.Sprintf("UnknownLevel(%d)", level)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (level *Level) UnmarshalText(text []byte) error {
	l, err := ParseLevel(string(text))
	if err != nil {
		return err
	}
	*level = Level(l)
	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (level Level) MarshalText() ([]byte, error) {
	switch level {
	case TraceLevel:
		return []byte("trace"), nil
	case DebugLevel:
		return []byte("debug"), nil
	case InfoLevel:
		return []byte("info"), nil
	case WarningLevel:
		return []byte("warning"), nil
	case ErrorLevel:
		return []byte("error"), nil
	case FatalLevel:
		return []byte("fatal"), nil
	case PanicLevel:
		return []byte("panic"), nil
	}
	return nil, fmt.Errorf("not a valid log level %d", level)
}

// ParseLevel parses the string as log level.
func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "panic":
		return PanicLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarningLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	case "trace":
		return TraceLevel, nil
	}
	return InfoLevel, fmt.Errorf("not a valid log level: %q", lvl)
}

// ParentLogger provides logger with logger factory that creates loggers with prefix.
type ParentLogger struct {
	Logger
	Prefix  string
	Factory LoggerFactory
}

// NewParentLogger creates new parent logger with given LoggerFactory and name as prefix.
func NewParentLogger(name string, factory LoggerFactory) *ParentLogger {
	return &ParentLogger{
		Logger:  factory.NewLogger(name),
		Prefix:  name,
		Factory: factory,
	}
}

// NewLogger returns logger using name prefixed with prefix defined in parent logger.
// If Factory is nil, DefaultRegistry is used.
func (p *ParentLogger) NewLogger(name string) Logger {
	factory := p.Factory
	if factory == nil {
		factory = DefaultRegistry
	}
	return factory.NewLogger(fmt.Sprintf("%s.%s", p.Prefix, name))
}

// PluginLogger is intended for:
// 1. small plugins (that just need one logger; name corresponds to plugin name)
// 2. large plugins that need multiple loggers (all loggers share same name prefix)
type PluginLogger interface {
	// Plugin has by default possibility to log
	// Logger name is initialized with plugin name
	Logger
	// LoggerFactory can be optionally used by large plugins
	// to create child loggers (their names are prefixed by plugin logger name)
	LoggerFactory
}

// ForPlugin is used to initialize plugin logger by name
// and optionally created children (their name prefixed by plugin logger name)
func ForPlugin(name string) PluginLogger {
	if logger, found := DefaultRegistry.Lookup(name); found {
		DefaultLogger.Tracef("using plugin logger for %q that was already initialized", name)
		return &ParentLogger{
			Logger:  logger,
			Prefix:  name,
			Factory: DefaultRegistry,
		}
	}
	return NewParentLogger(name, DefaultRegistry)
}
