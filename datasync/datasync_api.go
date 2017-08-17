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

package datasync

import (
	"io"

	"github.com/golang/protobuf/proto"
)

// TransportAdapter is an high-level abstraction of a transport to a remote
// data back end ETCD/Kafka/Rest/GRPC, used by Agent plugins to access data
// in a uniform & consistent way.
type TransportAdapter interface {
	Watcher
	Publisher
}

// Watcher is used by plugin to subscribe to both data change events and
// data resync events. Multiple keys can be specified, the caller will
// be subscribed to events on each key.
type Watcher interface {
	// WatchData using ETCD or any other data transport
	WatchData(resyncName string, changeChan chan ChangeEvent, resyncChan chan ResyncEvent,
		keyPrefixes ...string) (WatchDataRegistration, error)
}

// Publisher allows plugins to push their data changes to a data store.
type Publisher interface {
	// PublishData to ETCD or any other data transport (from other Agent Plugins)
	PublishData(key string, data proto.Message) error
}

// WatchDataRegistration is a facade that avoids importing the io.Closer package
// into Agent plugin implementations.
type WatchDataRegistration interface {
	io.Closer
}