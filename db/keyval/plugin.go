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

package keyval

import "github.com/ligato/cn-infra/core"

// KvPlugin provides unifying interface for different key-value datastore implementations.
type KvPlugin interface {
	core.Plugin
	// NewRootBroker returns a ProtoBroker providing methods for retrieving and editing key-value pairs in datastore.
	NewRootBroker() ProtoBroker
	// NewRootWatcher returns a ProtoWatcher providing for watching the changes in datastore.
	NewRootWatcher() ProtoWatcher
	// NewPrefixedBroker returns a ProtoBroker instance that prepends given keyPrefix to all keys in its calls.
	NewPrefixedBroker(keyPrefix string) ProtoBroker
	// NewPrefixedWatcher returns a ProtoWatcher instance. Given key prefix is prepended to keys during watch subscribe phase.
	// The prefix is removed from the key retrieved by GetKey() in ProtoWatchResp.
	NewPrefixedWatcher(keyPrefix string) ProtoWatcher
}

type kvPlugin struct {
	KvPlugin
}

// NewKvPlugin creates new instance of kv plugin with given key-val data store connection.
func NewKvPlugin(conn KvPlugin) KvPlugin {
	return &kvPlugin{conn}
}

func (k *kvPlugin) Init() error {
	return k.KvPlugin.Init()
}

func (k *kvPlugin) Close() error {
	return k.KvPlugin.Close()
}
