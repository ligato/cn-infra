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

package statuscheck

import (
	"context"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"

	"go.ligato.io/cn-infra/v2/agent"
	"go.ligato.io/cn-infra/v2/datasync"
	"go.ligato.io/cn-infra/v2/health/statuscheck/model/status"
	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/cn-infra/v2/logging"
)

var (
	// DefaultPublishPeriod is frequency of periodic writes of state data into ETCD.
	DefaultPublishPeriod = time.Second * 10
	// DefaultProbingPeriod is frequency of periodic plugin state probing.
	DefaultProbingPeriod = time.Second * 5
)

type Config struct {
	PublishPeriod time.Duration
	ProbingPeriod time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		PublishPeriod: DefaultPublishPeriod,
		ProbingPeriod: DefaultProbingPeriod,
	}
}

// Plugin struct holds all plugin-related data.
type Plugin struct {
	Deps

	conf *Config

	access sync.Mutex // lock for the Plugin data

	agentStat     *status.AgentStatus             // overall agent status
	interfaceStat *status.InterfaceStats          // interfaces' overall status
	pluginStat    map[string]*status.PluginStatus // plugin's status
	pluginProbe   map[string]PluginStateProbe     // registered status probes

	ctx    context.Context
	cancel context.CancelFunc // cancel can be used to cancel all goroutines and their jobs inside of the plugin
	wg     sync.WaitGroup     // wait group that allows to wait until all goroutines of the plugin have finished
}

// Deps lists the dependencies of statuscheck plugin.
type Deps struct {
	infra.PluginName                            // inject
	Log              logging.PluginLogger       // inject
	Transport        datasync.KeyProtoValWriter // inject (optional)
}

// Init prepares the initial status data.
func (p *Plugin) Init() error {
	p.Log.Debug("Init()")

	if p.conf == nil {
		p.conf = DefaultConfig()
	}

	// prepare initial status
	p.agentStat = &status.AgentStatus{
		State:        status.OperationalState_INIT,
		BuildVersion: agent.BuildVersion,
		BuildDate:    agent.BuildDate,
		CommitHash:   agent.CommitHash,
		StartTime:    time.Now().Unix(),
		LastChange:   time.Now().Unix(),
	}
	p.interfaceStat = &status.InterfaceStats{}
	p.pluginStat = make(map[string]*status.PluginStatus)
	p.pluginProbe = make(map[string]PluginStateProbe)

	// prepare context for all go routines
	p.ctx, p.cancel = context.WithCancel(context.Background())

	return nil
}

// AfterInit starts go routines for periodic probing and periodic updates.
// Initial state data are published via the injected transport.
func (p *Plugin) AfterInit() error {

	if err := p.StartProbing(); err != nil {
		return err
	}

	return nil
}

// Close stops go routines for periodic probing and periodic updates.
func (p *Plugin) Close() error {
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
	return nil
}

func (p *Plugin) StartProbing() error {
	// do periodic status probing for plugins that have provided the probe function
	p.wg.Add(1)
	go p.periodicProbing(p.ctx)

	// do periodic updates of the state data in ETCD
	p.wg.Add(1)
	go p.periodicUpdates(p.ctx)

	return p.publishInitial()
}

func (p *Plugin) publishInitial() error {
	p.access.Lock()
	defer p.access.Unlock()

	// transition to OK state if there are no plugins
	if len(p.pluginStat) == 0 {
		p.agentStat.State = status.OperationalState_OK
		p.agentStat.LastChange = time.Now().Unix()
	}
	if err := p.publishAgentData(); err != nil {
		p.Log.Warnf("publishing agent status failed: %v", err)
		return err
	} else {
		p.Log.Infof("Published agent status")
	}

	return nil
}

// GetAllPluginStatus returns a map containing pluginname and its status, for all plugins
func (p *Plugin) GetAllPluginStatus() map[string]*status.PluginStatus {
	//TODO - used currently, will be removed after incoporating improvements for exposing copy of map
	p.access.Lock()
	defer p.access.Unlock()

	return p.pluginStat
}

// GetInterfaceStats returns current global operational status of interfaces
func (p *Plugin) GetInterfaceStats() status.InterfaceStats {
	p.access.Lock()
	defer p.access.Unlock()

	return *p.interfaceStat
}

// GetAgentStatus return current global operational state of the agent.
func (p *Plugin) GetAgentStatus() status.AgentStatus {
	p.access.Lock()
	defer p.access.Unlock()
	return *p.agentStat
}

// Register a plugin for status change reporting.
func (p *Plugin) Register(pluginName infra.PluginName, probe PluginStateProbe) {
	if pluginName == "" {
		p.Log.Warnf("registering empty plugin name")
		return
	}

	stat := &status.PluginStatus{
		State:      status.OperationalState_INIT,
		LastChange: time.Now().Unix(),
	}

	p.access.Lock()
	defer p.access.Unlock()

	p.pluginStat[string(pluginName)] = stat
	if probe != nil {
		p.pluginProbe[string(pluginName)] = probe
	}

	p.Log.Tracef("Registered probe: %v", pluginName)

	// write initial status data into ETCD
	if err := p.publishPluginData(pluginName.String(), stat); err != nil {
		p.Log.Warnf("publishing plugin status failed: %v", err)
	}
}

// ReportStateChange can be used to report a change in the status of a previously registered plugin.
func (p *Plugin) ReportStateChange(pluginName infra.PluginName, state PluginState, lastError error) {
	p.reportStateChange(pluginName.String(), state, lastError)
}

// ReportStateChangeWithMeta can be used to report a change in the status of a previously registered plugin and report
// the specific metadata state
func (p *Plugin) ReportStateChangeWithMeta(pluginName infra.PluginName, state PluginState, lastError error, meta proto.Message) {
	p.reportStateChange(pluginName.String(), state, lastError)

	switch data := meta.(type) {
	case *status.InterfaceStats_Interface:
		p.reportInterfaceStateChange(data)
	default:
		p.Log.Debug("Unknown type of status metadata")
	}
}

func (p *Plugin) reportStateChange(pluginName string, state PluginState, lastError error) {
	p.access.Lock()
	defer p.access.Unlock()

	stat, ok := p.pluginStat[pluginName]
	if !ok {
		p.Log.Errorf("Unregistered plugin %s is reporting state, ignoring it.", pluginName)
		return
	}

	// update the state only if it has really changed
	changed := true
	if stateToProto(state) == stat.State {
		if lastError == nil && stat.Error == "" {
			changed = false
		}
		if lastError != nil && lastError.Error() == stat.Error {
			changed = false
		}
	}
	if !changed {
		return
	}

	p.Log.WithFields(logging.Fields{
		"plugin":  pluginName,
		"state":   state,
		"lastErr": lastError,
	}).Info("Agent plugin state update.")

	// update plugin state
	stat.State = stateToProto(state)
	stat.LastChange = time.Now().Unix()
	if lastError != nil {
		stat.Error = lastError.Error()
	} else {
		stat.Error = ""
	}
	if err := p.publishPluginData(pluginName, stat); err != nil {
		p.Log.Warnf("publishing plugin status failed: %v", err)
	}

	// update global state
	p.agentStat.State = stateToProto(state)
	p.agentStat.LastChange = time.Now().Unix()
	// Status for existing plugin
	var lastErr string
	if lastError != nil {
		lastErr = lastError.Error()
	}
	var pluginStatusExists bool
	for _, pluginStatus := range p.agentStat.Plugins {
		if pluginStatus.Name == pluginName {
			pluginStatusExists = true
			pluginStatus.State = stateToProto(state)
			pluginStatus.Error = lastErr
		}
	}
	// Status for new plugin
	if !pluginStatusExists {
		p.agentStat.Plugins = append(p.agentStat.Plugins, &status.PluginStatus{
			Name:  pluginName,
			State: stateToProto(state),
			Error: lastErr,
		})
	}
	if err := p.publishAgentData(); err != nil {
		p.Log.Warnf("publishing agent status failed: %v", err)
	}
}

func (p *Plugin) reportInterfaceStateChange(data *status.InterfaceStats_Interface) {
	p.access.Lock()
	defer p.access.Unlock()

	// Filter interfaces without internal name
	if data.InternalName == "" {
		p.Log.Debugf("Interface without internal name skipped for global status. Data: %v", data)
		return
	}

	// update only if state really changed
	var ifIndex int
	var existingData *status.InterfaceStats_Interface
	for index, ifState := range p.interfaceStat.Interfaces {
		// check if interface with the internal name already exists
		if data.InternalName == ifState.InternalName {
			ifIndex = index
			existingData = ifState
			break
		}
	}

	if existingData == nil {
		// new entry
		p.interfaceStat.Interfaces = append(p.interfaceStat.Interfaces, data)
		p.Log.Debugf("Global interface state data added: %v", data)
	} else if existingData.Index != data.Index || existingData.Status != data.Status || existingData.MacAddress != data.MacAddress {
		// updated entry - update only if state really changed
		p.interfaceStat.Interfaces = append(append(p.interfaceStat.Interfaces[:ifIndex], data), p.interfaceStat.Interfaces[ifIndex+1:]...)
		p.Log.Debugf("Global interface state data updated: %v", data)
	}
}

// publishAgentData writes the current global agent state into ETCD.
func (p *Plugin) publishAgentData() error {
	p.agentStat.LastUpdate = time.Now().Unix()
	if p.Transport != nil {
		return p.Transport.Put(status.AgentStatusKey(), p.agentStat)
	}
	return nil
}

// publishPluginData writes the current plugin state into ETCD.
func (p *Plugin) publishPluginData(pluginName string, pluginStat *status.PluginStatus) error {
	pluginStat.LastUpdate = time.Now().Unix()
	if p.Transport != nil {
		return p.Transport.Put(status.PluginStatusKey(pluginName), pluginStat)
	}
	return nil
}

// publishAll publishes global agent + all plugins state data into ETCD.
func (p *Plugin) publishAll() error {
	p.access.Lock()
	defer p.access.Unlock()

	if err := p.publishAgentData(); err != nil {
		return err
	}
	for name, s := range p.pluginStat {
		if err := p.publishPluginData(name, s); err != nil {
			return err
		}
	}
	return nil
}

// periodicProbing does periodic status probing for all plugins
// that have registered probe functions.
func (p *Plugin) periodicProbing(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-time.After(p.conf.ProbingPeriod):
			for pluginName, probe := range p.pluginProbe {
				state, err := probe()
				p.reportStateChange(pluginName, state, err)
				// just check in-between probes if the plugin is closing
				select {
				case <-ctx.Done():
					return
				default:
					continue
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

// periodicUpdates does periodic writes of state data into ETCD.
func (p *Plugin) periodicUpdates(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-time.After(p.conf.PublishPeriod):
			if err := p.publishAll(); err != nil {
				p.Log.Warnf("periodic status publishing failed: %v", err)
			}

		case <-ctx.Done():
			return
		}
	}
}

// stateToProto converts agent state type into protobuf agent state type.
func stateToProto(state PluginState) status.OperationalState {
	switch state {
	case Init:
		return status.OperationalState_INIT
	case OK:
		return status.OperationalState_OK
	default:
		return status.OperationalState_ERROR
	}
}
