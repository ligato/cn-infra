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
	"fmt"
	"sync"

	lg "github.com/sirupsen/logrus"

	"go.ligato.io/cn-infra/v2/logging"
)

// DefaultLoggerName is logger name of global instance of logger
const DefaultLoggerName = "defaultLogger"

var (
	defaultLogger = &Logger{
		name:   DefaultLoggerName,
		Logger: lg.StandardLogger(),
	}
)

func init() {
	lg.SetFormatter(DefaultFormatter)
	logging.DefaultLogger = defaultLogger
}

// DefaultLogger returns a global Logrus logger.
// Note, that recommended approach is to create a custom logger.
func DefaultLogger() *Logger {
	return defaultLogger
}

var _ logging.Logger = (*Logger)(nil)

// Logger is wrapper of Logrus logger. In addition to Logrus functionality it
// allows to define static log fields that are added to all subsequent log entries. It also automatically
// appends file name and line where the log is coming from. In order to distinguish logs from different
// go routines a tag (number that is based on the stack address) is computed. To achieve better readability
// numeric value of a tag can be replaced by a string using SetTag function.
type Logger struct {
	*lg.Logger

	name         string
	verbosity    int
	staticFields sync.Map
}

// NewLogger is a constructor creates instances of named logger.
// This constructor is called from LogRegistry which is useful
// when log levels needs to be changed by management API (such as REST)
//
// Example:
//
//    logger := NewLogger("loggerXY")
//    logger.Info()
//
func NewLogger(name string) *Logger {
	l := lg.New()
	logger := &Logger{
		Logger: l,
		name:   name,
	}
	logger.SetFormatter(DefaultFormatter)
	return logger
}

// GetName return the logger name
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
	var wasErr error
	staticFieldsMap := make(map[string]interface{})

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

	// throw panic outside of logger.Range()
	if wasErr != nil {
		panic(wasErr)
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

// WithField creates an entry from the standard logger and adds a field to
// it. If you want multiple fields, use `WithFields`.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the LogMsg it returns.
func (logger *Logger) WithField(key string, value interface{}) logging.LogWithLevel {
	return logger.withFields(logging.Fields{key: value}, 1)
}

// WithFields creates an entry from the standard logger and adds multiple
// fields to it. This is simply a helper for `WithField`, invoking it
// once for each field.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the LogMsg it returns.
func (logger *Logger) WithFields(fields logging.Fields) logging.LogWithLevel {
	return logger.withFields(fields, 1)
}

func (logger *Logger) withFields(fields logging.Fields, depth int) *Entry {
	static := logger.GetStaticFields()

	data := make(lg.Fields, len(fields)+len(static))
	for k, v := range static {
		data[k] = v
	}
	for k, v := range fields {
		data[k] = v
	}

	data[loggerKey] = logger.name

	return &Entry{
		logger:  logger,
		lgEntry: logger.Logger.WithFields(data),
	}
}

func (logger *Logger) header(depth int) *Entry {
	return logger.withFields(nil, 2)
}

// Debug logs a message at level Debug on the standard logger.
func (logger *Logger) log(lvl lg.Level, args ...interface{}) {
	if logger.IsLevelEnabled(lvl) {
		logger.header(1).Log(lvl, args...)
	}
}

func (logger *Logger) Log(lvl lg.Level, args ...interface{}) {
	if logger.IsLevelEnabled(lvl) {
		logger.header(1).Log(lvl, args...)
	}
}

func (logger *Logger) Logf(lvl lg.Level, f string, args ...interface{}) {
	if logger.IsLevelEnabled(lvl) {
		logger.header(1).Logf(lvl, f, args...)
	}
}

func (logger *Logger) Logln(lvl lg.Level, args ...interface{}) {
	if logger.IsLevelEnabled(lvl) {
		logger.header(1).Logln(lvl, args...)
	}
}

// Debug logs a message at level Debug on the standard logger.
func (logger *Logger) Debug(args ...interface{}) {
	if logger.Level >= lg.DebugLevel {
		logger.header(1).Debug(args...)
	}
}

// Info logs a message at level Info on the standard logger.
func (logger *Logger) Print(args ...interface{}) {
	logger.header(1).Info(args...)
}

// Info logs a message at level Info on the standard logger.
func (logger *Logger) Info(args ...interface{}) {
	if logger.Level >= lg.InfoLevel {
		logger.header(1).Info(args...)
	}
}

// Warn logs a message at level Warning on the standard logger.
func (logger *Logger) Warn(args ...interface{}) {
	if logger.Level >= lg.WarnLevel {
		logger.header(1).Warn(args...)
	}
}

// Warning logs a message at level Warn on the standard logger.
func (logger *Logger) Warning(args ...interface{}) {
	if logger.Level >= lg.WarnLevel {
		logger.header(1).Warning(args...)
	}
}

// Error logs a message at level Error on the standard logger.
func (logger *Logger) Error(args ...interface{}) {
	if logger.Level >= lg.ErrorLevel {
		logger.header(1).Error(args...)
	}
}

// Fatal logs a message at level Fatal on the standard logger.
func (logger *Logger) Fatal(args ...interface{}) {
	if logger.Level >= lg.FatalLevel {
		logger.header(1).Fatal(args...)
	}
}

// Panic logs a message at level Panic on the standard logger.
func (logger *Logger) Panic(args ...interface{}) {
	logger.header(1).Panic(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func (logger *Logger) Debugf(format string, args ...interface{}) {
	if logger.Level >= lg.DebugLevel {
		logger.header(1).Debugf(format, args...)
	}
}

// Printf logs a message at level Info on the standard logger.
func (logger *Logger) Printf(format string, args ...interface{}) {
	logger.header(1).Printf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func (logger *Logger) Infof(format string, args ...interface{}) {
	if logger.Level >= lg.InfoLevel {
		logger.header(1).Infof(format, args...)
	}
}

// Warnf logs a message at level Warn on the standard logger.
func (logger *Logger) Warnf(format string, args ...interface{}) {
	if logger.Level >= lg.WarnLevel {
		logger.header(1).Warnf(format, args...)
	}
}

// Warningf logs a message at level Warn on the standard logger.
func (logger *Logger) Warningf(format string, args ...interface{}) {
	if logger.Level >= lg.WarnLevel {
		logger.header(1).Warningf(format, args...)
	}
}

// Errorf logs a message at level Error on the standard logger.
func (logger *Logger) Errorf(format string, args ...interface{}) {
	if logger.Level >= lg.ErrorLevel {
		logger.header(1).Errorf(format, args...)
	}
}

// Fatalf logs a message at level Fatal on the standard logger.
func (logger *Logger) Fatalf(format string, args ...interface{}) {
	if logger.Level >= lg.FatalLevel {
		logger.header(1).Fatalf(format, args...)
	}
}

// Panicf logs a message at level Panic on the standard logger.
func (logger *Logger) Panicf(format string, args ...interface{}) {
	logger.header(1).Panicf(format, args...)
}

// Debugln logs a message at level Debug on the standard logger.
func (logger *Logger) Debugln(args ...interface{}) {
	if logger.Level >= lg.DebugLevel {
		logger.header(1).Debugln(args...)
	}
}

// Println logs a message at level Info on the standard logger.
func (logger *Logger) Println(args ...interface{}) {
	logger.header(1).Println(args...)
}

// Infoln logs a message at level Info on the standard logger.
func (logger *Logger) Infoln(args ...interface{}) {
	if logger.Level >= lg.InfoLevel {
		logger.header(1).Infoln(args...)
	}
}

// Warningln logs a message at level Warn on the standard logger.
func (logger *Logger) Warningln(args ...interface{}) {
	if logger.Level >= lg.WarnLevel {
		logger.header(1).Warningln(args...)
	}
}

// Errorln logs a message at level Error on the standard logger.
func (logger *Logger) Errorln(args ...interface{}) {
	if logger.Level >= lg.ErrorLevel {
		logger.header(1).Errorln(args...)
	}
}

// Panicln logs a message at level Panic on the standard logger.
func (logger *Logger) Panicln(args ...interface{}) {
	logger.header(1).Panicln(args...)
}

// Fatalln logs a message at level Fatal on the standard logger.
func (logger *Logger) Fatalln(args ...interface{}) {
	if logger.Level >= lg.FatalLevel {
		logger.header(1).Fatalln(args...)
	}
}
