//  Copyright (c) 2018 Cisco and/or its affiliates.
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

package agent

import (
	"context"
	"os"
	"reflect"
	"syscall"
	"time"

	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/infra"
)

var (
	// DefaultStartTimeout is default timeout for starting agent
	DefaultStartTimeout = time.Second * 15
	// DefaultStopTimeout is default timeout for stopping agent
	DefaultStopTimeout = time.Second * 5

	// DumpStackTraceOnTimeout prints stack trace on timeout or agent start/stop
	DumpStackTraceOnTimeout = os.Getenv("DUMP_STACK_ON_TIMEOUT") != ""
)

// Options specifies option list for the Agent
type Options struct {
	Name    string
	Version string

	Config config.Config
	//FlagSet *config.FlagSet
	Flags []config.Flag

	StartTimeout time.Duration
	StopTimeout  time.Duration
	QuitSignals  []os.Signal
	QuitChan     chan struct{}
	Context      context.Context

	Plugins []infra.Plugin

	initialized bool
}

func newOptions() Options {
	opt := Options{
		Name:         "agent",
		Version:      "dev",
		Config:       config.DefaultConfig,
		StartTimeout: DefaultStartTimeout,
		StopTimeout:  DefaultStopTimeout,
		QuitSignals:  []os.Signal{syscall.SIGINT, syscall.SIGTERM},
		Context:      context.Background(),
	}
	return opt
}

// Option is a function that operates on an Agent's Option
type Option func(*Options)

// Name sets agent name.
func Name(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

// Version sets agent version.
func Version(version string) Option {
	return func(o *Options) {
		o.Version = version
	}
}

func Flags(flags ...config.Flag) Option {
	return func(o *Options) {
		o.Flags = append(o.Flags, flags...)
	}
}

// StartTimeout returns an Option that sets timeout for the start of Agent.
func StartTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.StartTimeout = timeout
	}
}

// StopTimeout returns an Option that sets timeout for the stop of Agent.
func StopTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.StopTimeout = timeout
	}
}

// VersionInfo sets the version of the Agent to the entered string
func VersionInfo(buildVer, buildDate, commitHash string) Option {
	return func(o *Options) {
		BuildVersion = buildVer
		BuildDate = buildDate
		CommitHash = commitHash
	}
}

// Context returns an Option that sets the context for the Agent
func Context(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
	}
}

// QuitSignals returns an Option that will set signals which stop Agent
func QuitSignals(sigs ...os.Signal) Option {
	return func(o *Options) {
		o.QuitSignals = sigs
	}
}

// QuitOnClose returns an Option that will set channel which stops Agent on close
func QuitOnClose(ch chan struct{}) Option {
	return func(o *Options) {
		o.QuitChan = ch
	}
}

// Plugins creates an Option that adds a list of Plugins to the Agent's Plugin list
func Plugins(plugins ...infra.Plugin) Option {
	return func(o *Options) {
		o.Plugins = append(o.Plugins, plugins...)
	}
}

// AllPlugins creates an Option that adds all of the nested
// plugins recursively to the Agent's plugin list.
func AllPlugins(plugin infra.Plugin) Option {
	pluginMap := make(map[infra.Plugin]struct{})
	pluginNames := make(map[string]struct{})

	return func(o *Options) {
		typ := reflect.TypeOf(plugin)
		infraLogger.Debugf("searching for all deps in: %v (type: %v)", plugin, typ)

		foundPlugins, err := findPlugins(reflect.ValueOf(plugin), pluginMap)
		if err != nil {
			panic(err)
		}

		infraLogger.Debugf("found %d plugins in: %v (type: %v)", len(foundPlugins), plugin, typ)
		for _, plug := range foundPlugins {
			infraLogger.Debugf(" - plugin: %v (%v)", plug, reflect.TypeOf(plug))

			if _, ok := pluginNames[plug.String()]; ok {
				infraLogger.Fatalf("plugin with name %q already registered", plug.String())
			}
			pluginNames[plug.String()] = struct{}{}
		}
		o.Plugins = append(o.Plugins, foundPlugins...)

		// TODO: perhaps set plugin name to typ.String() if it's empty
		/*p, ok := plugin.(core.PluginNamed)
		if !ok {
			p = core.NamePlugin(typ.String(), plugin)
		}*/

		//pluginName := plugin.String()
		/*if plugin.String() == "" {
			if p, ok := plugin.(interface{ SetName(string) }); ok {
				infraLogger.Warnf("setting plugin name to: %q", typ.String())
				p.SetName(typ.String())
			}
		}*/

		if _, ok := pluginNames[plugin.String()]; ok {
			infraLogger.Fatalf("plugin with name %q already registered, custom name should be used", plugin.String())
		}
		pluginNames[plugin.String()] = struct{}{}
		o.Plugins = append(o.Plugins, plugin)
	}
}
