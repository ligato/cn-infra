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

package resync

import (
	"sync"
	"time"

	"github.com/ligato/cn-infra/logging"

	"github.com/ligato/cn-infra/logging/logroot"
)

var (
	gPlugin *Plugin
)

// plugin function is used in api to access the plugin instance. It panics if the plugin instance is no
func plugin() *Plugin {
	if gPlugin == nil {
		logroot.StandardLogger().Panic("Resync Orchestration is not yet initialized but you are trying to use that")
	}

	return gPlugin
}

// Plugin implements Plugin interface therefore can be loaded with other plugins
type Plugin struct {
	Log logging.Logger

	registrations map[string]Registration
	access        sync.Mutex
}

// Init initializes variables
func (plugin *Plugin) Init() (err error) {
	plugin.registrations = make(map[string]*Registration)

	//plugin.waingForResync = make(map[core.PluginName]*PluginEvent)
	//plugin.waingForResyncChan = make(chan *PluginEvent)
	//go plugin.watchWaingForResync()

	gPlugin = plugin

	return nil
}

// AfterInit method starts the resync
func (plugin *Plugin) AfterInit() (err error) {
	//plugin.startingResync()
	plugin.startResync()

	return nil
}

// Close TODO set flag that ignore errors => not start Resync while agent is stopping
// TODO kill existing Resync timeout while agent is stopping
func (plugin *Plugin) Close() error {
	//TODO close error report channel
	plugin.access.Lock()
	defer plugin.access.Unlock()
	plugin.registrations = make(map[string]*Registration)

	return nil
}

// Register function is supposed to be called in Init() by all VPP Agent plugins.
// The plugins are supposed to load current state of their objects when newResync() is called.
// But the actual CreateNewObjects(), DeleteObsoleteObjects() and ModifyExistingObjects() will be orchestrated
// to ensure there is proper order of that. If an error occurs during Resync than new Resync is planned.
func (plugin *Plugin) Register(resyncName string) Registration {
	plugin.access.Lock()
	defer plugin.access.Unlock()

	if _, found := plugin.registrations[resyncName]; found {
		plugin.Log.WithField("resyncName", resyncName).Panic("You are trying to register same resync twice")
		return nil
	}

	reg := NewRegistration(make(chan StatusEvent, 0)) /*Zero to have back pressure*/
	plugin.registrations[resyncName] = reg
	return reg
}

// call callback on plugins to create/delete/modify objects
func (plugin *Plugin) startResync() {
	for regName, reg := range plugin.registrations {
		started := newStatusEvent(Started)
		reg.StatusChan() <- started

		select {
		case <-started.ReceiveAck():
		case <-time.After(5 * time.Second):
			plugin.Log.WithField("regName", regName).Warn("Timeout of ACK")
		}
	}

	// TODO check if there ReportError (if not than report) if error occurred even during Resync
}
