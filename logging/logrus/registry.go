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
	"regexp"
	"sync"

	"github.com/sirupsen/logrus"

	"go.ligato.io/cn-infra/v2/logging"
)

func init() {
	logging.DefaultRegistry = DefaultRegistry()
}

var defaultRegistry = NewLogRegistry()

// DefaultRegistry returns the global registry instance.
func DefaultRegistry() *LogRegistry {
	return defaultRegistry
}

// NewLogRegistry is a constructor
func NewLogRegistry() *LogRegistry {
	registry := &LogRegistry{
		loggers:      new(sync.Map),
		logLevels:    make(map[string]logging.LogLevel),
		defaultLevel: initialLogLvl,
	}
	registry.putLoggerToMapping(defaultLogger)
	return registry
}

// LogRegistry contains logger map and rwlock guarding access to it
type LogRegistry struct {
	loggers      *sync.Map
	logLevels    map[string]logging.LogLevel
	defaultLevel logging.LogLevel
	hooks        []logrus.Hook
}

var validLoggerName = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`).MatchString

func checkLoggerName(name string) error {
	if !validLoggerName(name) {
		return fmt.Errorf("invalid logger name: %q, allowed only alphanum characters, dash and comma", name)
	}
	return nil
}

// NewLogger creates new named Logger instance. Name can be subsequently used to
// refer the logger in registry.
func (lr *LogRegistry) NewLogger(name string) logging.Logger {
	if existingLogger := lr.getLoggerFromMapping(name); existingLogger != nil {
		//panic(fmt.Errorf("logger with name '%s' already exists", name))
		fmt.Printf("logger with name '%s' already exists, returning previous logger\n", name)
		return existingLogger
	}
	if err := checkLoggerName(name); err != nil {
		panic(err)
	}

	logger := NewLogger(name)
	if lvl, ok := lr.logLevels[name]; ok {
		logger.SetLevel(lvl)
	} else {
		logger.SetLevel(lr.defaultLevel)
	}
	lr.putLoggerToMapping(logger)

	for _, hook := range lr.hooks {
		logger.AddHook(hook)
	}
	return logger
}

// ListLoggers returns a map (loggerName => log level)
func (lr *LogRegistry) ListLoggers() map[string]string {
	list := make(map[string]string)

	var wasErr error
	lr.loggers.Range(func(k, v interface{}) bool {
		key, ok := k.(string)
		if !ok {
			wasErr = fmt.Errorf("cannot cast log map key to string")
			return false
		}
		value, ok := v.(*Logger)
		if !ok {
			wasErr = fmt.Errorf("cannot cast log value to Logger obj")
			return false
		}
		list[key] = value.GetLevel().String()
		return true
	})
	// call panic outside of logger.Range()
	if wasErr != nil {
		panic(wasErr)
	}

	return list
}

// SetLevel modifies log level of selected logger in the registry
func (lr *LogRegistry) SetLevel(logger, level string) error {
	lvl, err := logging.ParseLogLevel(level)
	if err != nil {
		return err
	}
	if logger == "default" {
		lr.defaultLevel = lvl
		return nil
	}
	lr.logLevels[logger] = lvl
	logVal := lr.getLoggerFromMapping(logger)
	if logVal != nil {
		defaultLogger.Tracef("setting logger level: %v -> %v", logVal.name, lvl.String())
		logVal.SetLevel(lvl)
	}
	return nil
}

// GetLevel returns the currently set log level of the logger
func (lr *LogRegistry) GetLevel(logger string) (string, error) {
	logVal := lr.getLoggerFromMapping(logger)
	if logVal == nil {
		return "", fmt.Errorf("logger %s not found", logger)
	}
	return logVal.GetLevel().String(), nil
}

// Lookup returns a logger instance identified by name from registry
func (lr *LogRegistry) Lookup(loggerName string) (logger logging.Logger, found bool) {
	loggerInt, found := lr.loggers.Load(loggerName)
	if !found {
		return nil, false
	}
	logger, ok := loggerInt.(*Logger)
	if ok {
		return logger, found
	}
	panic(fmt.Errorf("cannot cast log value to Logger obj"))
}

// ClearRegistry removes all loggers except the default one from registry
func (lr *LogRegistry) ClearRegistry() {
	var wasErr error
	lr.loggers.Range(func(k, v interface{}) bool {
		key, ok := k.(string)
		if !ok {
			wasErr = fmt.Errorf("cannot cast log map key to string")
			return false
		}
		if key != globalName {
			lr.loggers.Delete(key)
		}
		return true
	})
	if wasErr != nil {
		panic(wasErr)
	}
}

// putLoggerToMapping writes logger into map of named loggers
func (lr *LogRegistry) putLoggerToMapping(logger *Logger) {
	lr.loggers.Store(logger.name, logger)
}

// getLoggerFromMapping returns a logger by its name
func (lr *LogRegistry) getLoggerFromMapping(logger string) *Logger {
	loggerVal, found := lr.loggers.Load(logger)
	if !found {
		return nil
	}
	log, ok := loggerVal.(*Logger)
	if ok {
		return log
	}
	panic("cannot cast log value to Logger obj")

}

// AddHook applies the hook to existing loggers and adds it to list of hooks
// to be applies for new loggers.
func (lr *LogRegistry) AddHook(hook logrus.Hook) {
	defaultLogger.Tracef("adding hook %q to log registry", hook)
	lr.hooks = append(lr.hooks, hook)
	for loggerName := range lr.ListLoggers() {
		logger, found := lr.lookupLogger(loggerName)
		if found {
			logger.AddHook(hook)
		}
	}
}

func (lr *LogRegistry) lookupLogger(name string) (*Logger, bool) {
	loggerInt, found := lr.loggers.Load(name)
	if !found {
		return nil, false
	}
	logger, ok := loggerInt.(*Logger)
	if ok {
		return logger, found
	}
	panic(fmt.Errorf("cannot cast log value to Logger obj"))
}
