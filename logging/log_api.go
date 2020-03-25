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

	"github.com/sirupsen/logrus"
)

// LogWithLevel allows to log with different log levels
type LogWithLevel interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infoln(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Warning(args ...interface{})
	Warningln(args ...interface{})
	Warningf(format string, args ...interface{})
	Error(args ...interface{})
	Errorln(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
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

// LogLevel represents severity of log record
type LogLevel = logrus.Level

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

// ParseLogLevel parses log level from string or returns InfoLevel on error.
func ParseLogLevel(str string) logrus.Level {
	if lvl, err := logrus.ParseLevel(str); err == nil {
		return lvl
	}
	return InfoLevel
}

// Fields is a type accepted by WithFields method. It can be used to instantiate map using shorter notation.
type Fields = logrus.Fields

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
		DefaultLogger.Debugf("using plugin logger for %q that was already initialized", name)
		return &ParentLogger{
			Logger:  logger,
			Prefix:  name,
			Factory: DefaultRegistry,
		}
	}
	return NewParentLogger(name, DefaultRegistry)
}
