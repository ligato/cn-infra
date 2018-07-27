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

package bolt

import (
	"github.com/boltdb/bolt"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/kvproto"
	"github.com/ligato/cn-infra/infra"
	"log"
	"os"
	"time"
)

// Config represents configuration for Bolt plugin.
type Config struct {
	DbPath            string        `json:"db-path"`
	FileMode          os.FileMode   `json:"file-mode"`
	LockTimeout       time.Duration `json:"lock-timeout"`
	SplitKeyToBuckets bool          `json:"split-key-to-buckets"`
}

// Plugin implements bolt plugin.
type Plugin struct {
	Deps

	// Plugin is disabled if there is no config file available
	disabled bool

	// Bolt DB encapsulation
	client *Client

	// Read/Write proto modelled data
	protoWrapper *kvproto.ProtoWrapper
}

// Deps lists dependencies of the etcd plugin.
// If injected, etcd plugin will use StatusCheck to signal the connection status.
type Deps struct {
	infra.Deps
}

// Disabled returns *true* if the plugin is not in use due to missing configuration.
func (plugin *Plugin) Disabled() bool {
	return plugin.disabled
}

// OnConnect executes callback from datasync
func (plugin *Plugin) OnConnect(callback func() error) {
	if err := callback(); err != nil {
		plugin.Log.Error(err)
	}
}

func (plugin *Plugin) getConfig() (*Config, error) {
	var cfg Config
	found, err := plugin.PluginConfig.GetValue(&cfg)
	if err != nil {
		return nil, err
	}
	if !found {
		plugin.Log.Info("Bolt config not found, skip loading this plugin")
		plugin.disabled = true
		return nil, nil
	}
	return &cfg, nil
}

// Init initializes Bolt plugin.
func (plugin *Plugin) Init() (err error) {
	cfg, err := plugin.getConfig()
	if err != nil || plugin.disabled {
		return err
	}

	plugin.client = &Client{}
	plugin.client.dbPath, err = bolt.Open(cfg.DbPath, cfg.FileMode, &bolt.Options{Timeout: cfg.LockTimeout})
	plugin.client.splitKeyToBuckets = cfg.SplitKeyToBuckets
	if err != nil {
		log.Fatal(err)
		plugin.disabled = true
		return err
	}

	plugin.protoWrapper = kvproto.NewProtoWrapperWithSerializer(plugin.client, &keyval.SerializerJSON{})

	plugin.Log.Infof("Bolt DB started %v", cfg.DbPath)
	return nil
}

// GetPluginName returns name of the plugin
func (plugin *Plugin) GetPluginName() infra.PluginName {
	return plugin.PluginName
}

// Close closes Bolt plugin.
func (plugin *Plugin) Close() error {
	if !plugin.disabled {
		plugin.client.dbPath.Close()
	}
	return nil
}

// NewBroker creates new instance of prefixed broker that provides API with arguments of type proto.Message.
func (plugin *Plugin) NewBroker(keyPrefix string) keyval.ProtoBroker {
	return plugin.protoWrapper.NewBroker(keyPrefix)
}

// NewWatcher creates new instance of prefixed broker that provides API with arguments of type proto.Message.
func (plugin *Plugin) NewWatcher(keyPrefix string) keyval.ProtoWatcher {
	return plugin.protoWrapper.NewWatcher(keyPrefix)
}
