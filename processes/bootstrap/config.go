// Copyright (c) 2019 Cisco and/or its affiliates.
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

package bootstrap

import "github.com/pkg/errors"

// Config represents configuration file for the bootstrap procedure
type Config struct {
	// The configuration file consists from a list of processes
	Processes []Process `json:"processes"`
}

// Process is a single process instance started by the bootstrap
type Process struct {
	// Unique name of the process
	Name string `json:"name"`

	// File where the log output will be written
	LogFilePath string `json:"log-file-path"`

	// Path to the process binary file
	BinaryPath string `json:"binary-path"`

	// A list of arguments the process will be started with
	Args []string `json:"args"`

	// This parameter allows to define another processes, which will be stopped
	// when the current instance is terminated (processes are defined by name).
	TriggerStopFor []string `json:"trigger-stop-for"`
}

// NewConf prepares a new empty configuration
func NewConf() *Config {
	return &Config{
		Processes: []Process{},
	}
}

// Get the configuration from a file
func (p *Plugin) getConfig() error {
	// in case config file was defined from outside
	if p.config != nil {
		return nil
	}
	p.config = NewConf()
	found, err := p.Cfg.LoadValue(p.config)
	if err != nil {
		return errors.Errorf("failed to load config file: %v", err)
	}
	if !found {
		return errors.Errorf("failed to load config file: not found")
	}

	return nil
}
