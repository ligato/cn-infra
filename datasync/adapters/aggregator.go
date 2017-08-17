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

package adapters

import (
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/persisted/dbsync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/datasync/rpc/grpcsync"
	"github.com/ligato/cn-infra/datasync/syncbase"
)

// TransportAggregator is cumulative adapter which contains all available transport types
type TransportAggregator struct {
	Adapters []datasync.TransportAdapter
}

// InitTransport initializes new transport with provided connection and stores it to the aggregator
func (ta *TransportAggregator) InitTransport(kvPlugin keyval.KvBytesPlugin, sl *servicelabel.Plugin, name string) {
	broker := kvPlugin.NewBroker(sl.GetAgentPrefix())
	watcher := kvPlugin.NewWatcher(sl.GetAgentPrefix())
	adapter := dbsync.NewAdapter(name, broker, watcher)
	ta.Adapters = append(ta.Adapters, adapter)
}

// InitGrpcTransport initializes a GRPC transport and stores it to the aggregator
func (ta *TransportAggregator) InitGrpcTransport() {
	grpcAdapter := grpcsync.NewAdapter()
	adapter := &syncbase.Adapter{Watcher: grpcAdapter}
	ta.Adapters = append(ta.Adapters, adapter)
}