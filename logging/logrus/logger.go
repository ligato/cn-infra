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

package logrus

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"

	"go.ligato.io/cn-infra/v2/logging"
)

const globalName = "global"

var initialLogLvl = logging.InfoLevel

func init() {
	if lvl, err := logging.ParseLogLevel(os.Getenv("INITIAL_LOGLVL")); err == nil {
		initialLogLvl = lvl
		defaultFormatter.Location = initialLogLvl >= logging.DebugLevel
		defaultFormatter.Function = initialLogLvl >= logging.TraceLevel
		defaultLogger.SetLevel(lvl)
		defaultLogger.Tracef("initial log level: %v", lvl.String())
	}
	logging.DefaultLogger = DefaultLogger()
}

var defaultLogger = WrapLogger(logrus.StandardLogger(), globalName)

// DefaultLogger returns a global logger.
// Note, that recommended approach is to create a custom logger.
func DefaultLogger() *Logger {
	return defaultLogger
}

// Logger is wrapper of Logrus logger. In addition to Logrus functionality it
// allows to define static log fields that are added to all subsequent log entries. It also automatically
// appends file name and line where the log is coming from. In order to distinguish logs from different
// go routines a tag (number that is based on the stack address) is computed. To achieve better readability
// numeric value of a tag can be replaced by a string using SetTag function.
type Logger struct {
	Logger *logrus.Logger

	name         string
	verbosity    int
	staticFields sync.Map
}

// WrapLogger wraps logrus.Logger and returns named Logger.
func WrapLogger(logger *logrus.Logger, name string) *Logger {
	logger.SetFormatter(DefaultFormatter())
	return &Logger{
		Logger: logger,
		name:   name,
	}
}

// NewLogger is a constructor creates instances of named logger.
// This constructor is called from LogRegistry which is useful
// when log levels needs to be changed by management API (such as REST)
//
// Example:
//
//    logger := NewLogger("MyLogger")
//    logger.Info("informative message")
//
func NewLogger(name string) *Logger {
	logger := WrapLogger(logrus.New(), name)
	return logger
}

func (logger *Logger) GetName() string {
	return logger.name
}

// SetStaticFields sets a map of fields that will be part of the each subsequent
// log entry of the logger
func (logger *Logger) SetStaticFields(fields map[string]interface{}) {
	for key, val := range fields {
		logger.staticFields.Store(key, val)
	}
}

// GetStaticFields returns currently set map of static fields - key-value pairs
// that are automatically added into log entry
func (logger *Logger) GetStaticFields() map[string]interface{} {
	staticFieldsMap := make(map[string]interface{})

	var wasErr error
	logger.staticFields.Range(func(k, v interface{}) bool {
		key, ok := k.(string)
		if !ok {
			wasErr = fmt.Errorf("cannot cast log map key to string")
			// false stops the iteration
			return false
		}
		staticFieldsMap[key] = v
		return true
	})
	if wasErr != nil {
		panic(wasErr) // call panic outside of logger.Range()
	}

	return staticFieldsMap
}

// SetVerbosity allows to set a logger verbosity. The verbosity can be used
// in custom loggers passed to external libraries (like GRPC) and may not
// correspond with the Logger plugin log levels. See the documentation of the
// given library to learn about supported verbosity levels.
func (logger *Logger) SetVerbosity(v int) {
	logger.verbosity = v
}

// V reports whether verbosity level is at least at the requested level
func (logger *Logger) V(l int) bool {
	return l <= logger.verbosity
}

func (logger *Logger) SetLevel(lvl logging.LogLevel) {
	logger.Logger.SetLevel(logrus.Level(lvl))
}

func (logger *Logger) GetLevel() logging.LogLevel {
	return logging.LogLevel(logger.Logger.GetLevel())
}

func (logger *Logger) AddHook(hook logrus.Hook) {
	logger.Logger.AddHook(hook)
}

func (logger *Logger) SetOutput(out io.Writer) {
	logger.Logger.SetOutput(out)
}

func (logger *Logger) SetFormatter(formatter logrus.Formatter) {
	logger.Logger.SetFormatter(formatter)
}

func (logger *Logger) SetReportCaller(enable bool) {
	logger.Logger.SetReportCaller(enable)
}

func (logger *Logger) WithError(err error) logging.LogWithLevel {
	return &Entry{
		logger:  logger,
		lgEntry: logger.entry().WithError(err),
	}
}

func (logger *Logger) WithContext(ctx context.Context) logging.LogWithLevel {
	return &Entry{
		logger:  logger,
		lgEntry: logger.entry().WithContext(ctx),
	}
}

// WithField creates an entry from the standard logger and adds a field to
// it. If you want multiple fields, use `WithFields`.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the LogMsg it returns.
func (logger *Logger) WithField(key string, value interface{}) logging.LogWithLevel {
	return logger.withFields(logging.Fields{key: value})
}

// WithFields creates an entry from the standard logger and adds multiple
// fields to it. This is simply a helper for `WithField`, invoking it
// once for each field.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the LogMsg it returns.
func (logger *Logger) WithFields(fields logging.Fields) logging.LogWithLevel {
	return logger.withFields(fields)
}

func (logger *Logger) withFields(fields logging.Fields) *Entry {
	return &Entry{
		logger:  logger,
		lgEntry: logger.entryWithFields(fields),
	}
}

func (logger *Logger) entryWithFields(fields logging.Fields) *logrus.Entry {
	static := logger.GetStaticFields()
	data := make(logrus.Fields, len(fields)+len(static)+1)
	for k, v := range static {
		data[k] = v
	}
	for k, v := range redactData(fields) {
		data[k] = v
	}
	data[LoggerKey] = logger.name
	return logger.Logger.WithFields(data)
}

func (logger *Logger) entry() *logrus.Entry {
	return logger.entryWithFields(nil)
}

func (logger *Logger) Log(lvl logrus.Level, args ...interface{}) {
	if logger.Logger.IsLevelEnabled(lvl) {
		logger.entry().Log(lvl, redactArgs(args)...)
	}
}

func (logger *Logger) Logf(lvl logrus.Level, f string, args ...interface{}) {
	if logger.Logger.IsLevelEnabled(lvl) {
		logger.entry().Log(lvl, fmt.Sprintf(f, redactArgs(args)...))
	}
}

func (logger *Logger) Logln(lvl logrus.Level, args ...interface{}) {
	if logger.Logger.IsLevelEnabled(lvl) {
		logger.entry().Log(lvl, sprintlnn(redactArgs(args)...))
	}
}

// Trace logs a message at level Trace on the standard logger.
func (logger *Logger) Trace(args ...interface{}) {
	logger.Log(logrus.TraceLevel, args...)
}

// Debug logs a message at level Debug on the standard logger.
func (logger *Logger) Debug(args ...interface{}) {
	logger.Log(logrus.DebugLevel, args...)
}

// Print logs a message at level Info on the standard logger.
func (logger *Logger) Print(args ...interface{}) {
	logger.Log(logrus.InfoLevel, args...)
}

// Info logs a message at level Info on the standard logger.
func (logger *Logger) Info(args ...interface{}) {
	logger.Log(logrus.InfoLevel, args...)
}

// Warn logs a message at level Warning on the standard logger.
func (logger *Logger) Warn(args ...interface{}) {
	logger.Log(logrus.WarnLevel, args...)
}

// Warning logs a message at level Warn on the standard logger.
func (logger *Logger) Warning(args ...interface{}) {
	logger.Log(logrus.WarnLevel, args...)
}

// Error logs a message at level Error on the standard logger.
func (logger *Logger) Error(args ...interface{}) {
	logger.Log(logrus.ErrorLevel, args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func (logger *Logger) Fatal(args ...interface{}) {
	logger.Log(logrus.FatalLevel, args...)
}

// Panic logs a message at level Panic on the standard logger.
func (logger *Logger) Panic(args ...interface{}) {
	logger.Log(logrus.PanicLevel, args...)
}

// Tracef logs a message at level Trae on the standard logger.
func (logger *Logger) Tracef(format string, args ...interface{}) {
	logger.Logf(logrus.TraceLevel, format, args...)
}

// Debugf logs a message at level Debug on the standard logger.
func (logger *Logger) Debugf(format string, args ...interface{}) {
	logger.Logf(logrus.DebugLevel, format, args...)
}

// Printf logs a message at level Info on the standard logger.
func (logger *Logger) Printf(format string, args ...interface{}) {
	logger.Logf(logrus.InfoLevel, format, args...)
}

// Infof logs a message at level Info on the standard logger.
func (logger *Logger) Infof(format string, args ...interface{}) {
	logger.Logf(logrus.InfoLevel, format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func (logger *Logger) Warnf(format string, args ...interface{}) {
	logger.Logf(logrus.WarnLevel, format, args...)
}

// Warningf logs a message at level Warn on the standard logger.
func (logger *Logger) Warningf(format string, args ...interface{}) {
	logger.Logf(logrus.WarnLevel, format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func (logger *Logger) Errorf(format string, args ...interface{}) {
	logger.Logf(logrus.ErrorLevel, format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func (logger *Logger) Fatalf(format string, args ...interface{}) {
	logger.Logf(logrus.FatalLevel, format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func (logger *Logger) Panicf(format string, args ...interface{}) {
	logger.Logf(logrus.PanicLevel, format, args...)
}

// Traceln logs a message at level Trace on the standard logger.
func (logger *Logger) Traceln(args ...interface{}) {
	logger.Logln(logrus.TraceLevel, args...)
}

// Debugln logs a message at level Debug on the standard logger.
func (logger *Logger) Debugln(args ...interface{}) {
	logger.Logln(logrus.DebugLevel, args...)
}

// Println logs a message at level Info on the standard logger.
func (logger *Logger) Println(args ...interface{}) {
	logger.Logln(logrus.InfoLevel, args...)
}

// Infoln logs a message at level Info on the standard logger.
func (logger *Logger) Infoln(args ...interface{}) {
	logger.Logln(logrus.InfoLevel, args...)
}

// Warnln logs a message at level Warn on the standard logger.
func (logger *Logger) Warnln(args ...interface{}) {
	logger.Logln(logrus.WarnLevel, args...)
}

// Warningln logs a message at level Warn on the standard logger.
func (logger *Logger) Warningln(args ...interface{}) {
	logger.Logln(logrus.WarnLevel, args...)
}

// Errorln logs a message at level Error on the standard logger.
func (logger *Logger) Errorln(args ...interface{}) {
	logger.Logln(logrus.ErrorLevel, args...)
}

// Fatalln logs a message at level Fatal on the standard logger.
func (logger *Logger) Fatalln(args ...interface{}) {
	logger.Logln(logrus.FatalLevel, args...)
}

// Panicln logs a message at level Panic on the standard logger.
func (logger *Logger) Panicln(args ...interface{}) {
	logger.Logln(logrus.PanicLevel, args...)
}
