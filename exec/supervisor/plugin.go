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

package supervisor

import (
	"os/exec"
	"sync"

	pm "github.com/ligato/cn-infra/exec/processmanager"
	"github.com/ligato/cn-infra/exec/processmanager/status"
	"github.com/ligato/cn-infra/infra"
	"github.com/pkg/errors"
)

const unnamedProcess = "<unnamed>"

// Supervisor defines methods to obtain information about running instances
type Supervisor interface {
	// GetProcessNames returns names of all running process instances
	GetProcessNames() []string
	// GetProcessByName returns an instance of given process
	GetProcessByName(name string) pm.ProcessInstance
}

// Plugin is a supervisor plugin structure
type Plugin struct {
	// a list of active process instances
	mx        sync.Mutex
	instances map[string]pm.ProcessInstance

	config *Config
	Deps
}

// Deps are supervisor dependencies (standard infra plugins + process manager)
type Deps struct {
	pm pm.ProcessManager
	infra.PluginDeps
}

// Init starts processes and their watchers defined by the configuration file
func (p *Plugin) Init() error {
	p.instances = make(map[string]pm.ProcessInstance)

	if err := p.getConfig(); err != nil {
		return err
	}

	// run watchers even if there are failed processes. They will be handled
	// as terminated and exit immediately
	err := p.start(p.config.Processes)
	p.watch()

	return err
}

// Close stops all remaining processes
func (p *Plugin) Close() error {
	p.mx.Lock()
	defer p.mx.Unlock()

	// TODO option to keep them running?

	for name, instance := range p.instances {
		if _, err := instance.StopAndWait(); err != nil {
			p.Log.Errorf("supervisor close error: failed to stop process %s: %v", name, err)
		}
	}

	return nil
}

// GetProcessNames returns names of all running process instances
func (p *Plugin) GetProcessNames() (names []string) {
	p.mx.Lock()
	defer p.mx.Unlock()

	for name := range p.instances {
		names = append(names, name)
	}
	return names
}

// GetProcessByName returns an instance of given process
func (p *Plugin) GetProcessByName(reqName string) pm.ProcessInstance {
	p.mx.Lock()
	defer p.mx.Unlock()

	for name, instance := range p.instances {
		if name == reqName {
			return instance
		}
	}
	return nil
}

func (p *Plugin) start(processes []Process) error {
	pmLogManager := newPmLoggerManager()

	var inError []string
	for _, process := range processes {
		if err := p.validate(&process); err != nil {
			p.Log.Errorf("unable to start process %s: %v", process.Name, err)
			inError = append(inError, process.Name)
			continue
		}
		processStat := make(chan status.ProcessStatus)
		pLogger, err := pmLogManager.newPmLogger(process.Name, process.LogFilePath)
		if err != nil {
			p.Log.Errorf("error preparing process logger %s: %v", process.Name, err)
			inError = append(inError, process.Name)
			continue
		}
		instance := p.pm.NewProcess(process.Name, process.BinaryPath,
			pm.Args(process.Args...),
			pm.Writer(pLogger, pLogger),
			pm.Notify(processStat),
			pm.AutoTerminate())
		if err := instance.Start(); err != nil {
			p.Log.Errorf("error starting process %s: %v", process.Name, err)
			inError = append(inError, process.Name)
			continue
		}
		p.instances[process.Name] = instance
		p.Log.Debugf("process %s started", process.Name)
	}

	if len(inError) != 0 {
		return errors.Errorf("following processes end up with error: %v", inError)
	}
	return nil
}

func (p *Plugin) validate(process *Process) error {
	if process.Name == "" {
		// so it will not be logged as empty space
		process.Name = unnamedProcess
		return errors.Errorf("name not defined")
	}
	if _, ok := p.instances[process.Name]; ok {
		return errors.Errorf("name %s already exists", process.Name)
	}
	if process.BinaryPath == "" {
		return errors.Errorf("binary %s path not defined", process.Name)
	}
	return nil
}

func (p *Plugin) watch() {
	for name, instance := range p.instances {
		// at first check whether the process is still alive
		if !instance.IsAlive() {
			// handle the process as if it was terminated
			p.Log.Warnf("%s does not run, handle as terminated", name)
			p.handleTerminated(name, instance)
		} else {
			go p.watchProcess(name, instance)
		}
		p.Log.Debug("process %s watcher started", name)
	}
}

func (p *Plugin) watchProcess(name string, instance pm.ProcessInstance) {
	for {
		stat := <-instance.GetNotificationChan()
		if stat == status.Terminated {
			p.Log.Infof("%s terminated", name)
			p.handleTerminated(name, instance)
			return
		}
	}
}

func (p *Plugin) handleTerminated(name string, instance pm.ProcessInstance) {
	p.mx.Lock()
	defer p.mx.Unlock()

	// find original config
	var process Process
	for _, processCfg := range p.config.Processes {
		if name == processCfg.Name {
			process = processCfg
			break
		}
	}

	// execute hooks
	for _, hook := range process.Hooks {
		if hook.Command != "" {
			if _, err := exec.Command(hook.Command, hook.CmdArgs...).Output(); err != nil {
				p.Log.Errorf("Failed to exec command %s, args %s: %v", hook.Command, hook.CmdArgs, err)
			}
		}
	}

	// stop all other processes if defined
	if process.Required {
		for name, instance := range p.instances {
			if instance.IsAlive() {
				if _, err := instance.StopAndWait(); err != nil {
					p.Log.Errorf("Attempt to stop %s failed: %v", name, err)
				}
			} else {
				p.Log.Debugf("Process %s is no longer running", name)
			}
		}
	}

	delete(p.instances, name)
}
