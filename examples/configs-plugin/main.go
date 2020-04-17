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

package main

import (
	"encoding/json"
	"time"

	"go.ligato.io/cn-infra/v2/agent"
	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/cn-infra/v2/logging"
)

// ExampleCfg defines config values as struct.
type ExampleCfg struct {
	Name        string
	Description string
	Port        uint
	Settings    *struct {
		Max     int
		Timeout time.Duration
	}
}

func main() {
	const PluginName = "example"

	p := &ExamplePlugin{
		Deps: Deps{
			PluginName: infra.PluginName(PluginName),
			Log:        logging.ForPlugin(PluginName),
			Config:     config.ForPlugin(PluginName),
		},
		exampleFinished: make(chan struct{}),
	}

	a := agent.NewAgent(
		agent.AllPlugins(p),
		agent.QuitOnClose(p.exampleFinished),
	)

	if err := a.Run(); err != nil {
		logging.Fatalf("Run() error: %+v", err)
	}
}

// ExamplePlugin demonstrates the use of injected Config plugin.
type ExamplePlugin struct {
	Deps

	Conf *ExampleCfg

	exampleFinished chan struct{}
}

// Deps defines dependencies for ExamplePlugin.
type Deps struct {
	infra.PluginName
	Log    logging.PluginLogger
	Config config.PluginConfig
}

// Init loads the configuration file assigned to ExamplePlugin (can be changed
// via the example-config flag).
// Loaded config is printed into the log file.
func (p *ExamplePlugin) Init() (err error) {
	p.Log.Debug("Loading plugin config ", p.Config.GetConfigName())

	p.Conf = &ExampleCfg{
		Name:        "defaultName",
		Description: "no description",
		Port:        9191,
	}
	p.Log.Infof("DEFAULT: %+v", p.Conf)

	found, err := p.Config.LoadValue(p.Conf)
	if err != nil {
		p.Log.Error("Error loading config", err)
	} else if !found {
		p.Log.Warn("Config not found, using default values")
	} else {
		p.Log.Info("Config loaded, values from", p.Config.GetConfigName())
	}
	p.Log.Infof("LOADED: %+v", p.Conf)

	b, err := json.MarshalIndent(p.Conf, "", "  ")
	if err != nil {
		return err
	}
	p.Log.Infof("CONFIG JSON: %s", b)

	go func() {
		time.Sleep(time.Second)
		close(p.exampleFinished)
	}()

	return nil
}

// Close closes the plugin.
func (p *ExamplePlugin) Close() (err error) {
	return nil
}
