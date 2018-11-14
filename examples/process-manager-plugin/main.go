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

// Test application just keeps running indefinitely, or for given time is defined
// via parameter creating a running process. The purpose is to serve as a test
// application for process manager example.

package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/process"
	"github.com/ligato/cn-infra/process/status"
	"github.com/pkg/errors"
)

const pluginName = "process-manager-example"

func main() {
	pmPlugin := process.DefaultPlugin
	example := &PMExample{
		Log:      logging.ForPlugin(pluginName),
		PM:       &pmPlugin,
		finished: make(chan struct{}),
	}

	a := agent.NewAgent(
		agent.AllPlugins(example),
		agent.QuitOnClose(example.finished),
	)
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

// PMExample demonstrates the usage of the process manager plugin.
type PMExample struct {
	Log logging.PluginLogger
	PM  process.API

	finished chan struct{}
}

// Init starts the example
func (p *PMExample) Init() error {
	go p.runExample()
	return nil
}

// Close frees the plugin resources.
func (p *PMExample) Close() error {
	return nil
}

// String returns name of the plugin.
func (p *PMExample) String() string {
	return pluginName
}

// Runs example with step by step description
func (p *PMExample) runExample() {
	// Step 1: simple process handling (start, restart, watch, delete)
	if err := p.simpleExample(); err != nil {
		p.Log.Errorf("simple process manager example failed with error: %v", err)
		close(p.finished)
		return
	}
	// Step 2: advanced process handing (attach running process, read status file)
	if err := p.advancedExample(); err != nil {
		p.Log.Errorf("simple process manager example failed with error: %v", err)
		close(p.finished)
		return
	}
	// Step 3: process templates (create a template, start process using template)
	if err := p.templateExample(); err != nil {
		p.Log.Errorf("simple process manager example failed with error: %v", err)
		close(p.finished)
		return
	}
	close(p.finished)
	return
}

func (p *PMExample) simpleExample() error {
	p.Log.Infof("1. starting simple process manager example")
	// Process manager plugin has internal cache for all processes created via its API. These processes
	// are considered as known to plugin. The example uses test-process - a simple application which
	// keeps running after start, so it can be handled by process manager.
	// At first, the process needs to be defined with unique name and command. Since test process allows
	// to use argument, use option 'Args()' to start the application with it and set max uptime to 60 seconds.
	// Then initialize status channel since status notifications are required, and provide it to the plugin
	// via 'Notify()'.
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	cmd := filepath.Join(currentDir, "test-process", "test-process")
	notifyChan := make(chan status.ProcessStatus)
	pr := p.PM.NewProcess("test-pr", cmd, process.Args("-max-uptime=60"), process.Notify(notifyChan))

	// The process is initialized. We can verify that via plugin manager API (prInst == pr)
	prInst := p.PM.GetProcessByName("test-pr")
	if prInst == nil {
		return errors.Errorf("expected process instance is nil")
	}

	// Since the watch channel is used, start watcher in another goroutine. The watcher will track the process status
	// during the example.
	var state status.ProcessStatus
	go func() {
		var ok bool
		for {
			select {
			case state, ok = <-notifyChan:
				if !ok {
					p.Log.Infof("===>(watcher) process watcher ended")
					return
				}
				p.Log.Infof("===>(watcher) received test process state: %s", state)
			}
		}
	}()
	if err := pr.Start(); err != nil {
		return err
	}
	p.Log.Infof("Let's wait for the process to start")
	time.Sleep(2 * time.Second)
	if state == status.Sleeping || state == status.Running || state == status.Idle {
		p.Log.Infof("success!")
	} else {
		return errors.Errorf("failed to start the test process within timeout")
	}

	// Now the process is running and we can use the instance to read various process definitions. The most common:
	pid := pr.GetPid()
	p.Log.Infof("PID: %d", pid)
	p.Log.Infof("Instance name: %s", pr.GetInstanceName())
	p.Log.Infof("Start time: %d", pr.GetStartTime().Nanosecond())
	p.Log.Infof("Uptime (s): %f", pr.GetUptime().Seconds())

	// Let's try to restart the process
	if err := pr.Restart(); err != nil {
		return err
	}
	p.Log.Infof("Let's wait for process to restart")
	time.Sleep(2 * time.Second)

	// Restarted process is expected to have different process ID. We stored old PID, so let's compare it with
	// the new one.
	if pid == pr.GetPid() {
		p.Log.Warnf("PID of restarted process is the same, perhaps the restart was not successful")
	} else {
		p.Log.Infof("success!")
	}
	p.Log.Infof("new PID: %d", pid)
	p.Log.Infof("Uptime (s): %f", pr.GetUptime().Seconds())

	// Now lets stop the process
	if err := pr.Stop(); err != nil {
		return err
	}
	p.Log.Infof("Let's wait for process to stop")
	time.Sleep(2 * time.Second)

	// Important: we stopped the process using SIGTERM, which causes process to become defunct because with the current
	// setup, the example is a parent process which is still running. If we try to start it again with 'Start()',
	// the operating system spawns a new process which instance will NOT be passed to the plugin manager by os package
	// and the new process will become unmanageable.
	// There are several options what to do:
	// * use Wait(). It waits for process status and terminates it completely.
	// * stop the process with the StopAndWait() which merges the two procedures together
	// * create a process with 'AutoTermination' option. It causes that every managed process which becomes zombie
	// will be terminated automatically.
	// Since we already stopped the process, the only thing we can do is to wait
	if _, err := pr.Wait(); err != nil {
		return err
	}
	p.Log.Infof("Let's wait for process to complete")
	time.Sleep(2 * time.Second)

	// Stopped process can be started again, etc. If we want to get rid of the process completely, we have to delete
	// it via plugin API. It requires plugin name (not instance name). Deleted process is not stopped if running.
	// Delete also closes notification channel.
	prName := pr.GetName()
	if err := p.PM.Delete(prName); err != nil {
		return err
	}
	if err := p.PM.GetProcessByName(prName); err != nil {
		return errors.Errorf("process was expected to be removed, but is still exists")
	}

	p.Log.Infof("simple process manager example is completed")

	return nil
}

func (p *PMExample) advancedExample() error {
	p.Log.Infof("2. starting advanced process manager example")
	// Let's prepare the application as before with some additional options which are 'Detach' and 'Restarts'.
	// Spawned process is by default a child process of the caller. It means that the child process will be
	// automatically terminated together with the parent. Option 'Detach' allows to detach the process from
	// parent and keeps is running.
	// Option 'Restarts' defines a number of automatic restarts if given process is terminated.
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	cmd := filepath.Join(currentDir, "test-process", "test-process")
	notifyChan := make(chan status.ProcessStatus)
	pr := p.PM.NewProcess("test-pr", cmd, process.Args("-max-uptime=60"), process.Notify(notifyChan),
		process.Detach(), process.Restarts(1))

	// Start the watcher as before and ensure the process is running
	var state status.ProcessStatus
	go func() {
		var ok bool
		for {
			select {
			case state, ok = <-notifyChan:
				if !ok {
					p.Log.Infof("===>(watcher-old) process watcher ended")
					return
				}
				p.Log.Infof("===>(watcher-old) received test process state: %s", state)
			}
		}
	}()
	if err := pr.Start(); err != nil {
		return err
	}
	p.Log.Infof("Let's wait for process to start")
	time.Sleep(2 * time.Second)
	if state == status.Sleeping || state == status.Running || state == status.Idle {
		p.Log.Infof("success!")
	} else {
		return errors.Errorf("failed to start the test process within timeout")
	}

	// The example cannot simulate parent restart, so we use delete to 'forget' the process and keep it running. But first,
	// we need to remember the process ID
	pid := pr.GetPid()
	p.Log.Infof("PID: %d", pid)

	// Make sure the process is known to plugin
	prInst := p.PM.GetProcessByName(pr.GetName())
	if prInst == nil {
		return errors.Errorf("expected process instance is nil")
	}

	// Now delete the process
	p.Log.Infof("Deleting process...")
	if err := p.PM.Delete(pr.GetName()); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)

	// And make sure the process is NOT known to plugin
	prInst = p.PM.GetProcessByName(pr.GetName())
	if prInst != nil {
		return errors.Errorf("process expected to be removed still exists")
	}

	// Since we know the PID, the plugin can reattach to the same instance it created with a new name. From
	// the plugin perspective, attaching the process is just another way of creating it, so all the options have to be
	// re-defined. Note: it is possible to attach to process without command or arguments, but it is not possible
	// to start such a process instance. The notify channel need to be re-initialized and new watcher started, because
	// the previous one was closed by the Delete().
	p.Log.Infof("Reattaching process...")
	notifyChan = make(chan status.ProcessStatus)
	go func() {
		var ok bool
		for {
			select {
			case state, ok = <-notifyChan:
				if !ok {
					p.Log.Infof("===>(watcher-new) process watcher ended")
					return
				}
				p.Log.Infof("===>(watcher-new) received test process state: %s", state)
			}
		}
	}()
	if pr, err = p.PM.AttachProcess("test-pr-attached", cmd, pid, process.Args("-max-uptime=60"), process.Notify(notifyChan),
		process.Detach(), process.Restarts(1)); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)

	// Make sure the process is known again to plugin
	prInst = p.PM.GetProcessByName(pr.GetName())
	if prInst == nil {
		return errors.Errorf("expected process instance is nil")
	}
	p.Log.Infof("success!")
	p.Log.Infof("reattached PID: %d", pid)

	// Since the restart count is set to 1, process will be restarted if terminated. It does not matter how, so let's
	// just stop it.
	if _, err := pr.StopAndWait(); err != nil {
		return err
	}
	p.Log.Infof("Let's wait while the process is stopped and restarted")
	time.Sleep(3 * time.Second)
	if state == status.Sleeping || state == status.Running || state == status.Idle {
		p.Log.Infof("success!")
	} else {
		return errors.Errorf("failed to re-start the test process within timeout")
	}
	// Process was stopped and auto-started again with new PID
	p.Log.Infof("PID after auto-restart: %d", pr.GetPid())

	// Every process creates a status file within /proc/<pid>/status with a plenty of information about the process
	// state, CPU or memory usage, etc. The process watcher periodically reads the status data for current state to
	// propagate changes. To read status, use ReadStatus().
	prStatus, err := pr.ReadStatus(pr.GetPid())
	if err != nil {
		return err
	}
	// Some example status data
	p.Log.Infof("Threads: %d", prStatus.Threads)
	p.Log.Infof("Allowed CPUs: %s", prStatus.CpusAllowed)
	p.Log.Infof("Parent process ID: %d", prStatus.PPid)
	p.Log.Infof("Total program size: %s", prStatus.VMSize)

	// Stop and delete the process. It will not be run again, since maximum of restarts was set to 1.
	p.Log.Infof("Stopping and removing the process...")
	if _, err := pr.StopAndWait(); err != nil {
		return err
	}
	if err := p.PM.Delete(pr.GetName()); err != nil {
		return err
	}
	p.Log.Infof("done")
	p.Log.Infof("advanced process manager example is completed")

	return nil
}

func (p *PMExample) templateExample() error {
	p.Log.Infof("3. starting template example")

	// Template example requires to have a template path defined in process manager config, and this config needs to be
	// provided to the example
	if _, err := p.PM.GetAllTemplates(); err != nil {
		p.Log.Warnf("template example aborted, no config file was provided for the process manager plugin")
	}

	// A template represents a process configuration defined by the model inside the process manager plugin.
	// The template allows to define all the setup items from previous examples (arguments, watcher, restarts, etc.).
	// The plugin API defines a method NewProcessFromTemplate(<template>). The template object can be programmed
	// manually as a *process.Template object. But main reason for templates to exist is that they are stored
	// in the filesystem as JSON objects, thus can persist an application restart.
	// To create a template file, the proto model has to be used. However template can be also generated from
	// new/attached process with options. There is an option 'Template' allowing it.
	// Lets create a new process with template. The template file will be created in path defined in plugin config.
	// The option has a run-on-startup parameter which can be set to true. If so, all the template processes will be
	// also started in the process manager init phase and available as soon as the application using it is loaded.
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	cmd := filepath.Join(currentDir, "test-process", "test-process")
	notifyChan := make(chan status.ProcessStatus)
	pr := p.PM.NewProcess("test-pr", cmd, process.Args("-max-uptime=60"), process.Notify(notifyChan),
		process.Template(false))

	// Start the watcher as before and ensure the process is running
	var state status.ProcessStatus
	go func() {
		var ok bool
		for {
			select {
			case state, ok = <-notifyChan:
				if !ok {
					p.Log.Infof("===>(watcher-process) process watcher ended")
					return
				}
				p.Log.Infof("===>(watcher-process) received test process state: %s", state)
			}
		}
	}()
	if err := pr.Start(); err != nil {
		return err
	}
	p.Log.Infof("Let's wait for process to start")
	time.Sleep(2 * time.Second)
	if state == status.Sleeping || state == status.Running || state == status.Idle {
		p.Log.Infof("success!")
	} else {
		return errors.Errorf("failed to start the test process within timeout")
	}

	// Lets verify the JSON template file exists
	template, err := p.PM.GetTemplate(pr.GetName())
	if err != nil {
		return err
	}
	if template == nil {
		return errors.Errorf("expected template does not exists")
	}
	p.Log.Infof("template for test process was created")

	// Now we have the process template, so lets stop and remove the running process, in order to start it again with it
	p.Log.Infof("terminating running process...")
	if _, err := pr.StopAndWait(); err != nil {
		return err
	}
	prName := pr.GetName()
	if err := p.PM.Delete(prName); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	if prInst := p.PM.GetProcessByName(prName); prInst != nil {
		return errors.Errorf("expected terminated process instance is still running")
	}

	// Re-crate the plugin-process instance using template file and start it as usually with new watcher
	pr = p.PM.NewProcessFromTemplate(template)
	go func() {
		var ok bool
		for {
			select {
			case state, ok = <-pr.GetNotificationChan():
				if !ok {
					p.Log.Infof("===>(watcher-template) process watcher ended")
					return
				}
				p.Log.Infof("===>(watcher-template) received test process state: %s", state)
			}
		}
	}()
	if err := pr.Start(); err != nil {
		return err
	}
	p.Log.Infof("Let's wait for template process to start")
	time.Sleep(2 * time.Second)
	if state == status.Sleeping || state == status.Running || state == status.Idle {
		p.Log.Infof("success!")
	} else {
		return errors.Errorf("failed to start the test process within timeout")
	}

	p.Log.Infof("template example finished")
	return nil
}
