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

package statusidx

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/health/statuscheck/model/status"
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/idxmap/mem"
	"github.com/ligato/cn-infra/logging/logroot"
	log "github.com/ligato/cn-infra/logging/logrus"
)

// PluginStatusIdxMap provides read-only access to mapping between software interface indexes (used internally in VPP)
// and interface names.
type PluginStatusIdxMap interface {
	// GetMapping returns internal read-only mapping with Value of type interface{}.
	GetMapping() idxmap.NamedMapping

	// GetValue looks up previously stored item identified by index in mapping.
	GetValue(name string) (data *status.PluginStatus, exists bool)

	// LookupByState returns name of items that contains given state
	LookupByState(ip string) []string

	// WatchNameToIdx allows to subscribe for watching changes in pluginStatusIdxMap mapping
	WatchNameToIdx(subscriber core.PluginName, pluginChannel chan PluginStatusEvent)
}

// PluginStatusIdxMapRW is mapping between software interface indexes (used internally in VPP)
// and interface names.
type PluginStatusIdxMapRW interface {
	PluginStatusIdxMap

	// RegisterName adds new item into name-to-index mapping.
	Put(name string, pluginStatus *status.PluginStatus)

	// UnregisterName removes an item identified by name from mapping
	Delete(name string) (data *status.PluginStatus, exists bool)
}

// NewPluginStatusIdxMap is a constructor
func NewPluginStatusIdxMap(owner core.PluginName) PluginStatusIdxMap {
	return &pluginStatusIdxMap{mapping: mem.NewNamedMapping(logroot.Logger(), owner, IndexPluginStatus)}
}

// pluginStatusIdxMap is type-safe implementation of mapping between Software interface index
// and interface name. It holds as well Value of type *InterfaceMeta.
type pluginStatusIdxMap struct {
	mapping idxmap.NamedMappingRW
}

// PluginStatusEvent represents an item sent through watch channel in pluginStatusIdxMap.
// In contrast to NameToIdxDto it contains typed Value.
type PluginStatusEvent struct {
	idxmap.NamedMappingEvent
	Value *status.PluginStatus
}

const (
	stateIndexKey = "stateKey"
)

// GetMapping returns internal read-only mapping. It is used in tests to inspect the content of the pluginStatusIdxMap.
func (swi *pluginStatusIdxMap) GetMapping() idxmap.NamedMapping {
	return swi.mapping
}

// RegisterName adds new item into name-to-index mapping.
func (swi *pluginStatusIdxMap) Put(name string, pluginStatus *status.PluginStatus) {
	swi.mapping.Put(name, pluginStatus)
}

// IndexPluginStatus creates indexes for Value. Index for State will be created
func IndexPluginStatus(data interface{}) map[string][]string {
	log.Debug("IndexPluginStatus ", data)

	indexes := map[string][]string{}
	pluginStatus, ok := data.(*status.PluginStatus)
	if !ok || pluginStatus == nil {
		return indexes
	}

	state := pluginStatus.State
	if state != 0 {
		indexes[stateIndexKey] = []string{state.String()}
	}
	return indexes
}

// UnregisterName removes an item identified by name from mapping
func (swi *pluginStatusIdxMap) Delete(name string) (data *status.PluginStatus, exists bool) {
	meta, exists := swi.mapping.Delete(name)
	return swi.castdata(meta), exists
}

// GetValue looks up previously stored item identified by index in mapping.
func (swi *pluginStatusIdxMap) GetValue(name string) (data *status.PluginStatus, exists bool) {
	meta, exists := swi.mapping.GetValue(name)
	if exists {
		data = swi.castdata(meta)
	}
	return data, exists
}

// LookupNameByIP returns names of items that contains given IP address in Value
func (swi *pluginStatusIdxMap) LookupByState(state status.OperationalState) []string {
	return swi.mapping.ListNames(stateIndexKey, state.String())
}

func (swi *pluginStatusIdxMap) castdata(meta interface{}) *status.PluginStatus {
	if pluginStatus, ok := meta.(*status.PluginStatus); ok {
		return pluginStatus
	}

	return nil
}

// WatchNameToIdx allows to subscribe for watching changes in pluginStatusIdxMap mapping
func (swi *pluginStatusIdxMap) WatchNameToIdx(subscriber core.PluginName, pluginChannel chan PluginStatusEvent) {
	swi.mapping.Watch(subscriber, func(event idxmap.NamedMappingGenericEvent) {
		pluginChannel <- PluginStatusEvent{
			NamedMappingEvent: event.NamedMappingEvent,
			Value:             swi.castdata(event.Value),
		}
	})
}
