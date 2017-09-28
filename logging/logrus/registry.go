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

	"github.com/Sirupsen/logrus"
	"github.com/ligato/cn-infra/logging"
	"sync"
	"unsafe"
	"sync/atomic"
)

// NewLogRegistry is a constructor
func NewLogRegistry() logging.Registry {
	registry := &logRegistry{}
	// Init mapping in loggers
	registry.loggers.Store(make(map[string]*Logger))
	// Put default logger loggers
	registry.put(defaultLogger)
	return registry
}

// logRegistry contains logger map and rwlock guarding access to it
type logRegistry struct {
	access    sync.RWMutex
	// loggers holds mapping of logger instances indexed by their names
	mapping map[string] /*logger name*/ *Logger
	loggers atomic.Value
}

// NewLogger creates new named Logger instance. Name can be subsequently used to
// refer the logger in registry.
func (lr *logRegistry) NewLogger(name string) logging.Logger {
	existingLogger := lr.getLoggerFromMap(name)
	if existingLogger != nil {
		panic(fmt.Errorf("logger with name '%s' already exists", name))
	}
	if err := checkLoggerName(name); err != nil {
		panic(err)
	}

	logger := NewLogger(name)

	lr.put(logger)
	return logger
}

// ListLoggers returns a map (loggerName => log level)
func (lr *logRegistry) ListLoggers() map[string]string {
	mapping := lr.loggers.Load().(map[string]*Logger)
	list := map[string]string{}
	for k, v := range mapping {
		list[k] = v.GetLevel().String()
	}
	return list
}

// SetLevel modifies log level of selected logger in the registry
func (lr *logRegistry) SetLevel(logger, level string) error {
	mapping := lr.loggers.Load().(map[string]*Logger)
	lg, ok := mapping[logger]
	if !ok {
		return fmt.Errorf("logger %s not found", logger)
	}
	lvl, err := logrus.ParseLevel(level)
	if err == nil {
		switch lvl {
		case logrus.DebugLevel:
			lg.SetLevel(logging.DebugLevel)
		case logrus.InfoLevel:
			lg.SetLevel(logging.InfoLevel)
		case logrus.WarnLevel:
			lg.SetLevel(logging.WarnLevel)
		case logrus.ErrorLevel:
			lg.SetLevel(logging.ErrorLevel)
		case logrus.PanicLevel:
			lg.SetLevel(logging.PanicLevel)
		case logrus.FatalLevel:
			lg.SetLevel(logging.FatalLevel)
		}

	}

	return nil
}

// GetLevel returns the currently set log level of the logger
func (lr *logRegistry) GetLevel(logger string) (string, error) {
	mapping := lr.loggers.Load().(map[string]*Logger)
	lg, ok := mapping[logger]
	if !ok {
		return "", fmt.Errorf("logger %s not found", logger)
	}
	return lg.GetLevel().String(), nil
}

// Lookup returns a logger instance identified by name from registry
func (lr *logRegistry) Lookup(loggerName string) (logger logging.Logger, found bool) {
	mapping := lr.loggers.Load().(map[string]*Logger)
	logger, found = mapping[loggerName]
	return
}

// ClearRegistry removes all loggers except the default one from registry
func (lr *logRegistry) ClearRegistry() {
	mapping := lr.loggers.Load().(map[string]*Logger)
	for k := range mapping {
		if k != DefaultLoggerName {
			delete(mapping, k)
		}
	}
	lr.loggers.Store(mapping)
}

// put writes logger into map of named loggers
func (lr *logRegistry) put(logger *Logger) {
	mapping := lr.loggers.Load().(map[string]*Logger)
	mapping[logger.name] = logger

	lr.loggers.Store(mapping)
}

func (lr *logRegistry) getLoggerFromMap(logger string) *Logger {
	lr.access.RLock()
	defer lr.access.RUnlock()

	loggerObj, ok := lr.mapping[logger]
	if ok {
		unsafeLogger := (*unsafe.Pointer)(unsafe.Pointer(&loggerObj))
		return (*Logger)(atomic.LoadPointer(unsafeLogger))

	}
	return nil
}

func (lr *logRegistry) storeLoggerToMap(logger string, newVal *Logger) *Logger {
	lr.access.RLock()
	defer lr.access.RUnlock()

	loggerObj, ok := lr.mapping[logger]
	if ok {
		unsafeLogger := (*unsafe.Pointer)(unsafe.Pointer(&loggerObj))
		loggerVal := (*Logger)(atomic.LoadPointer(unsafeLogger))
		old := loggerVal
		atomic.CompareAndSwapPointer(unsafeLogger, unsafe.Pointer(old), unsafe.Pointer(loggerVal))

	}
	return nil
}
