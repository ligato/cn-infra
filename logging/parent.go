//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package logging

import "fmt"

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
	Logger
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
