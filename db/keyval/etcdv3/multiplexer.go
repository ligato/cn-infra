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

package etcdv3

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/utils/safeclose"
)

// DbMuxEtcd implements Plugin interface therefore can be loaded with other plugins
type DbMuxEtcd struct {
	client *clientv3.Client
	dbRoot *ProtoBrokerEtcd
}

// NewEtcdPlugin creates a new instance of DbMuxEtcd. Configuration of etcd connection is loaded from file.
func NewEtcdPlugin(etcdConfig string) (*DbMuxEtcd, error) {
	client, err := initRemoteClient(etcdConfig)
	if err != nil {
		return nil, err
	}
	return NewEtcdPluginUsingClient(client), nil
}

// NewEtcdPluginUsingClient creates a new instance of DbMuxEtcd using given etcd client
func NewEtcdPluginUsingClient(client *clientv3.Client) *DbMuxEtcd {
	return &DbMuxEtcd{client: client}
}

// Init is entry point called by Agent Core
func (plugin *DbMuxEtcd) Init() (err error) {
	/*etcdLogger, err := log_registry.NewLogger("EtcdLib")
	if err != nil {
		log.Error(err)
		return err
	}
	etcdv3.SetLogger(etcdLogger)*/

	coreBroker, err := NewBytesBrokerUsingClient(plugin.client)
	if err != nil {
		log.Error(err)
		return err
	}
	plugin.dbRoot = NewProtoBrokerWithSerializer(coreBroker, &keyval.SerializerJSON{})

	log.Debug("initEtcdClient success ", coreBroker)
	return nil
}

// NewRootBroker provides methods for retrieving and editing key-value pairs in the etcd
func (plugin *DbMuxEtcd) NewRootBroker() keyval.ProtoBroker {
	return plugin.dbRoot
}

// NewRootWatcher provides methods for watching for changes in the Data Store
func (plugin *DbMuxEtcd) NewRootWatcher() keyval.ProtoWatcher {
	return plugin.dbRoot.NewPluginBroker("")
}

// NewPrefixedBroker returns a ProtoBroker instance that prepends given keyPrefix to all keys in its calls.
func (plugin *DbMuxEtcd) NewPrefixedBroker(keyPrefix string) keyval.ProtoBroker {
	return plugin.dbRoot.NewPluginBroker(keyPrefix)
}

// NewPrefixedWatcher returns a ProtoWatcher instance. Given keyprefix is prepended to keys during watch subscribe phase.
// The prefix is removed from the key retrieved by GetKey() in ProtoWatchResp.
func (plugin *DbMuxEtcd) NewPrefixedWatcher(keyPrefix string) keyval.ProtoWatcher {
	return plugin.dbRoot.NewPluginBroker(keyPrefix)
}

// Close cleans up the resources
func (plugin *DbMuxEtcd) Close() error {
	return safeclose.Close(plugin.dbRoot)
}
