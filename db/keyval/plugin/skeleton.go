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

package plugin

import (
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/utils/safeclose"
	"io"
)

// Connection defines an access to a particular key-value data store implementation.
type Connection interface {
	NewBroker(keyPrefix string) keyval.ProtoBroker
	NewWatcher(keyPrefix string) keyval.ProtoWatcher
	io.Closer
}

// Skeleton of a KV plugin is a generic part of KV plugin.
type Skeleton struct {
	conn    Connection
	connect func() (Connection, error)
}

// NewSkeleton creates a new instance of the Skeleton with the given connector.
// The connection is established in AfterInit phase.
func NewSkeleton(connector func() (Connection, error)) *Skeleton {
	return &Skeleton{connect: connector}
}

// Init is called on plugin startup
func (plugin *Skeleton) Init() (err error) {
	return nil
}

// AfterInit is called once all plugin have been initialized. The connection to datastore
// is established in this phase.
func (plugin *Skeleton) AfterInit() (err error) {
	plugin.conn, err = plugin.connect()
	return err
}

// Close cleans up the resources
func (plugin *Skeleton) Close() error {
	return safeclose.Close(plugin.conn)
}
