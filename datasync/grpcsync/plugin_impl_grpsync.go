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

package grpcsync

import (
	"sync"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/syncbase"
	"github.com/ligato/cn-infra/rpc/grpc"
)

// Plugin grpcsync implements Plugin interface, therefore can be loaded with other plugins.
type Plugin struct {
	Deps
	Adapter  datasync.KeyValProtoWatcher
	initOnce sync.Once
}

// Deps - gRPC Plugin dependencies
type Deps struct {
	GRPC grpc.Server
}

// Init registers new gRPC service and instantiates plugin.Adapter.
func (plugin *Plugin) Init() (err error) {
	plugin.initOnce.Do(func() {
		grpcAdapter := NewAdapter(plugin.GRPC.GetServer())
		plugin.Adapter = &syncbase.Adapter{Watcher: grpcAdapter}
	})
	return err
}

// Close does nothing.
func (plugin *Plugin) Close() error {
	// Warning: If you ever do anything here other than return nil, please see grpc plugin for an example of how to
	// Use closeOnce (a sync.Once) to protect it.
	return nil
}
