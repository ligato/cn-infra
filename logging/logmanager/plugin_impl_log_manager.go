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

package logmanager

import (
	"fmt"
	"github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/evalphobia/logrus_fluent"
	"github.com/gorilla/mux"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/sirupsen/logrus"
	lgSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/unrolled/render"
	"log/syslog"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// LoggerData encapsulates parameters of a logger represented as strings.
type LoggerData struct {
	Logger string `json:"logger"`
	Level  string `json:"level"`
}

// Variable names in logger registry URLs
const (
	loggerVarName = "logger"
	levelVarName  = "level"
)

// Plugin allows to manage log levels of the loggers.
type Plugin struct {
	Deps

	*Conf
}

// Deps groups dependencies injected into the plugin so that they are
// logically separated from other plugin fields.
type Deps struct {
	infra.Deps
	LogRegistry logging.Registry  // inject
	HTTP        rest.HTTPHandlers // inject
}

// NewConf creates default configuration with InfoLevel & empty loggers.
// Suitable also for usage in flavor to programmatically specify default behavior.
func NewConf() *Conf {
	return &Conf{
		DefaultLevel: "",
		Loggers:      []ConfLogger{},
	}
}

// Conf is a binding that supports to define default log levels for multiple loggers
type Conf struct {
	DefaultLevel string       `json:"default-level"`
	Loggers      []ConfLogger `json:"loggers"`

	// logging hooks configuration
	Hooks map[string]HookConfig
}

// ConfLogger is configuration of a particular logger.
// Currently we support only logger level.
type ConfLogger struct {
	Name  string
	Level string //debug, info, warning, error, fatal, panic
}

// Init does nothing
func (lm *Plugin) Init() error {
	if lm.PluginConfig != nil {
		if lm.Conf == nil {
			lm.Conf = NewConf()
		}

		lm.Log.Debugf("logs config: %+v", lm.Conf)
		_, err := lm.PluginConfig.GetValue(lm.Conf)
		if err != nil {
			return err
		}

		// Handle default log level. Prefer value from environmental variable
		defaultLogLvl := os.Getenv("INITIAL_LOGLVL")
		if defaultLogLvl == "" {
			defaultLogLvl = lm.Conf.DefaultLevel
		}
		if defaultLogLvl != "" {
			if err := lm.LogRegistry.SetLevel("default", defaultLogLvl); err != nil {
				lm.Log.Warnf("setting default log level failed: %v", err)
			} else {
				// All loggers created up to this point were created with initial log level set (defined
				// via INITIAL_LOGLVL env. variable with value 'info' by default), so at first, let's set default
				// log level for all of them.
				for loggerName := range lm.LogRegistry.ListLoggers() {
					logger, exists := lm.LogRegistry.Lookup(loggerName)
					if !exists {
						continue
					}
					logger.SetLevel(stringToLogLevel(defaultLogLvl))
				}
			}
		}

		// Handle config file log levels
		for _, logCfgEntry := range lm.Conf.Loggers {
			// Put log/level entries from configuration file to the registry.
			if err := lm.LogRegistry.SetLevel(logCfgEntry.Name, logCfgEntry.Level); err != nil {
				// Intentionally just log warn & not propagate the error (it is minor thing to interrupt startup)
				lm.Log.Warnf("setting log level %s for logger %s failed: %v", logCfgEntry.Level,
					logCfgEntry.Name, err)
			}
		}
		lm.Log.Warn("configuring log hooks ...")

		if hookConfig, exists := lm.Conf.Hooks[HookSysLog]; exists {
			lm.AddHook(HookSysLog, hookConfig)
		}
		if hookConfig, exists := lm.Conf.Hooks[HookLogStash]; exists {
			lm.AddHook(HookLogStash, hookConfig)
		}
		if hookConfig, exists := lm.Conf.Hooks[HookFluent]; exists {
			lm.AddHook(HookFluent, hookConfig)
		}
	}

	return nil
}

// AfterInit is called at plugin initialization. It register the following handlers:
// - List all registered loggers:
//   > curl -X GET http://localhost:<port>/log/list
// - Set log level for a registered logger:
//   > curl -X PUT http://localhost:<port>/log/<logger-name>/<log-level>
func (lm *Plugin) AfterInit() error {
	if lm.HTTP != nil {
		lm.HTTP.RegisterHTTPHandler(fmt.Sprintf("/log/{%s}/{%s:debug|info|warning|error|fatal|panic}",
			loggerVarName, levelVarName), lm.logLevelHandler, "PUT")
		lm.HTTP.RegisterHTTPHandler("/log/list", lm.listLoggersHandler, "GET")
	}
	return nil
}

// Close is called at plugin cleanup phase.
func (lm *Plugin) Close() error {
	return nil
}

// ListLoggers lists all registered loggers.
func (lm *Plugin) listLoggers() []LoggerData {
	var loggers []LoggerData

	lgs := lm.LogRegistry.ListLoggers()
	for lg, lvl := range lgs {
		ld := LoggerData{
			Logger: lg,
			Level:  lvl,
		}
		loggers = append(loggers, ld)
	}

	return loggers
}

// setLoggerLogLevel modifies the log level of the all loggers in a plugin
func (lm *Plugin) setLoggerLogLevel(name string, level string) error {
	lm.Log.Debugf("SetLogLevel name '%s', level '%s'", name, level)

	return lm.LogRegistry.SetLevel(name, level)
}

// logLevelHandler processes requests to set log level on loggers in a plugin
func (lm *Plugin) logLevelHandler(formatter *render.Render) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		lm.Log.Infof("Path: %s", req.URL.Path)
		vars := mux.Vars(req)
		if vars == nil {
			formatter.JSON(w, http.StatusNotFound, struct{}{})
			return
		}
		err := lm.setLoggerLogLevel(vars[loggerVarName], vars[levelVarName])
		if err != nil {
			formatter.JSON(w, http.StatusNotFound,
				struct{ Error string }{err.Error()})
			return
		}
		formatter.JSON(w, http.StatusOK,
			LoggerData{Logger: vars[loggerVarName], Level: vars[levelVarName]})
	}
}

// listLoggersHandler processes requests to list all registered loggers
func (lm *Plugin) listLoggersHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		formatter.JSON(w, http.StatusOK, lm.listLoggers())
	}
}

// convert log level string representation to DebugLevel value
func stringToLogLevel(level string) logging.LogLevel {
	level = strings.ToLower(level)
	switch level {
	case "debug":
		return logging.DebugLevel
	case "info":
		return logging.InfoLevel
	case "warn":
		return logging.WarnLevel
	case "error":
		return logging.ErrorLevel
	case "fatal":
		return logging.FatalLevel
	case "panic":
		return logging.PanicLevel
	}

	return logging.InfoLevel
}

const (
	HookSysLog   = "syslog"
	HookLogStash = "logstash"
	HookFluent   = "fluent"
)

type HookConfig struct {
	Protocol string
	Address  string
	Port     int
	Levels   []string
}

type CommonHook struct {
	logrus.Hook
	levels      []logrus.Level
}

func (cH *CommonHook) Levels() []logrus.Level {
	return cH.levels
}

func (lm *Plugin) AddHook(hookName string, hookConfig HookConfig) error {
	var cHook *CommonHook
	var lgHook logrus.Hook
	var err error

	switch hookName {
	case HookSysLog:
		var address = hookConfig.Address
		if hookConfig.Address != "" {
			address = address + ":" + strconv.Itoa(hookConfig.Port)
		}
		lgHook, err = lgSyslog.NewSyslogHook(
			hookConfig.Protocol,
			address,
			syslog.LOG_INFO,
			"")
	case HookLogStash:
		lgHook, err = logrustash.NewHook(
			hookConfig.Protocol,
			hookConfig.Address+":"+strconv.Itoa(hookConfig.Port),
			"vpp-agent")
	case HookFluent:
		lgHook, err = logrus_fluent.NewWithConfig(logrus_fluent.Config{
			Host: hookConfig.Address,
			Port: hookConfig.Port,
		})
	default:
		return fmt.Errorf("unsupported hook")
	}
	// create hook
	cHook = &CommonHook{
		lgHook,
		[]logrus.Level{},
	}
	// fill up defined levels, or use default if not defined
	if len(hookConfig.Levels) == 0 {
		lgl, _ := logrus.ParseLevel(lm.Conf.DefaultLevel)
		cHook.levels = append(cHook.levels, lgl)
	} else {
		for _, level := range hookConfig.Levels {
			lgl, _ := logrus.ParseLevel(level)
			cHook.levels = append(cHook.levels, lgl)
		}
	}
	// add hook to existing loggers and store it into registry for late use
	if err == nil {
		lgs := lm.LogRegistry.ListLoggers()
		for lg, _ := range lgs {
			logger, found := lm.LogRegistry.Lookup(lg)
			if found {
				logger.AddHook(cHook)
			}
		}
		lm.Log.Warnf("add hook %v to registry", hookName)
		lm.LogRegistry.AddHook(cHook)
	} else {
		lm.Log.Warnf("couldn't create hook for %v : %v", hookName, err.Error())
	}
	return err
}
