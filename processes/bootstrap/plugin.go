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

import (
	"github.com/ligato/cn-infra/infra"
	pm "github.com/ligato/cn-infra/processes/processmanager"
	"github.com/ligato/cn-infra/processes/processmanager/status"
	"github.com/pkg/errors"
	"os"
	"sync"
)

type Bootstrap interface {
	GetProcessNames() []string
	GetProcessByName(name string) pm.ProcessInstance
}

// Plugin is a bootstrap plugin structure
type Plugin struct {
	mx sync.Mutex

	// a list of active process instances
	instances map[string]pm.ProcessInstance

	config *Config
	Deps
}

// Deps are bootstrap dependencies (standard infra plugins + process manager)
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

	p.start(p.config.Processes)
	p.watch()

	return nil
}

// Close stops all remaining processes
func (p *Plugin) Close() error {
	p.mx.Lock()
	defer p.mx.Unlock()

	// TODO option to keep them running?

	for name, instance := range p.instances {
		if _, err := instance.StopAndWait(); err != nil {
			p.Log.Errorf("bootstrap close error: failed to stop process %s: %v", name, err)
		}
	}

	return nil
}

// String is a plugin name
func (p *Plugin) String() string {
	return p.PluginName.String()
}

func (p *Plugin) GetProcessNames() (names []string) {
	for name := range p.instances {
		names = append(names, name)
	}
	return names
}

func (p *Plugin) GetProcessByName(reqName string) pm.ProcessInstance {
	for name, instance := range p.instances {
		if name == reqName {
			return instance
		}
	}
	return nil
}

func (p *Plugin) start(processes []Process) {
	for _, process := range processes {
		if err := p.validate(&process); err != nil {
			p.Log.Errorf("unable to start process %s: %v", process.Name, err)
			continue
		}
		processStat := make(chan status.ProcessStatus)
		pLogger := newPmLogger(process.Name, process.LogFilePath)
		instance := p.pm.NewProcess(process.Name, process.BinaryPath, pm.Args(process.Args...),
			pm.Writer(pLogger, pLogger), pm.Notify(processStat), pm.AutoTerminate())
		if err := instance.Start(); err != nil {
			p.Log.Errorf("Error starting process %s: %v", process.Name, err)
			os.Exit(-1)
		}
		p.instances[process.Name] = instance
		p.Log.Debugf("process %s started", process.Name)
	}
}

func (p *Plugin) validate(process *Process) error {
	if process.Name == "" {
		return errors.Errorf("name not defined")
	}
	if _, ok := p.instances[process.Name]; ok {
		return errors.Errorf("name already exists", process.Name)
	}
	if process.BinaryPath == "" {
		return errors.Errorf("binary path not defined", process.Name)
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
	for _, toStop := range process.TriggerStopFor {
		toStopInstance, ok := p.instances[toStop]
		if !ok {
			p.Log.Warnf("Non-existing process %s requested to stop", toStop)
			continue
		}
		if toStopInstance.IsAlive() {
			if _, err := toStopInstance.StopAndWait(); err != nil {
				p.Log.Errorf("Attempt to stop %s failed: %v", toStop, err)
			}
		} else {
			p.Log.Info("Process %s is no longer running", toStop)
		}
	}
	delete(p.instances, name)
	if len(p.instances) == 0 {
		p.Log.Info("No more processes are running, exiting")
		os.Exit(0)
	}
}
































