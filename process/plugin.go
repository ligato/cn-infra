// Copyright (c) 2018 Cisco and/or its affiliates.
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

package process

import (
	"os"

	"github.com/ligato/cn-infra/process/status"

	"github.com/ligato/cn-infra/infra"
	"github.com/pkg/errors"
)

// API defines methods to create, delete or manage processes
type API interface {
	// NewProcess creates new process instance with name, command to start and other options (arguments, policy).
	// New process is not immediately started, process instance comprises from a set of methods to manage.
	NewProcess(cmd string, options ...POption) ManagerAPI
	// Starts process from template file
	NewProcessFromTemplate(path string) error
	// Attach to existing process using its process ID. The process is stored under the provided name. Error
	// is returned if process does not exits
	AttachProcess(pid int, options ...POption) (ManagerAPI, error)
	// GetProcessByName returns existing process instance using name
	GetProcessByName(name string) ManagerAPI
	// GetProcessByName returns existing process instance using PID
	GetProcessByPID(pid int) ManagerAPI
	// GetAll returns all processes known to plugin
	GetAll() []ManagerAPI
}

// Plugin implements API to manage processes. There are two options to add a process to manage, start it as a new one
// or attach to an existing process. In both cases, the process is stored internally as known to the plugin.
type Plugin struct {
	processes []*Process

	Deps
}

// Deps define process dependencies
type Deps struct {
	infra.PluginDeps
}

// Init does nothing for process manager plugin
func (p *Plugin) Init() error {
	p.Log.Debugf("Initializing process manager plugin")

	return nil
}

// Close does nothing for process manager plugin
func (p *Plugin) Close() error {
	return nil
}

// String returns string representation of the plugin
func (p *Plugin) String() string {
	return p.PluginName.String()
}

// AttachProcess attaches to existing process and reads its status
func (p *Plugin) AttachProcess(pid int, options ...POption) (ManagerAPI, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, errors.Errorf("cannot attach to process with PID %d: %v", pid, err)
	}
	attachedPr := &Process{
		log:        p.Log,
		process:    process,
		sh:         &status.Reader{Log: p.Log},
		notifChan:  make(chan status.ProcessStatus),
		cancelChan: make(chan struct{}),
	}
	for _, option := range options {
		option(attachedPr.options)
	}
	p.processes = append(p.processes, attachedPr)

	attachedPr.status, err = attachedPr.sh.ReadStatus(attachedPr.GetPid())
	if err != nil {
		p.Log.Warnf("failed to read process (PID %d) status: %v", pid, err)
	}

	go attachedPr.watch()

	return attachedPr, nil
}

// NewProcess creates new process
func (p *Plugin) NewProcess(cmd string, options ...POption) ManagerAPI {
	newPr := &Process{
		log:        p.Log,
		cmd:        cmd,
		options:    &POptions{},
		sh:         &status.Reader{Log: p.Log},
		status:     &status.File{},
		notifChan:  make(chan status.ProcessStatus),
		cancelChan: make(chan struct{}),
	}
	for _, option := range options {
		option(newPr.options)
	}
	p.processes = append(p.processes, newPr)

	go newPr.watch()

	return newPr
}

func (p *Plugin) NewProcessFromTemplate(path string) error {
	return nil
}

// GetProcess uses process name to find a desired instance
func (p *Plugin) GetProcessByName(name string) ManagerAPI {
	for _, process := range p.processes {
		if process.status.Name == name {
			return process
		}
	}
	return nil
}

// GetProcess uses process ID to find a desired instance
func (p *Plugin) GetProcessByPID(pid int) ManagerAPI {
	for _, process := range p.processes {
		if process.status.Pid == pid {
			return process
		}
	}
	return nil
}

// GetAll returns all processes known to plugin
func (p *Plugin) GetAll() []ManagerAPI {
	var processes []ManagerAPI
	for _, process := range p.processes {
		processes = append(processes, process)
	}
	return processes
}
