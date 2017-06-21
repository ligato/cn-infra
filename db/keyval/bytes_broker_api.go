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

// BytesBroker allows to store, retrieve and remove data in a key-value form
type BytesBroker interface {
	// Put puts single key-value pair into etcd
	Put(key string, data []byte) error
	// NewTxn creates a transaction
	NewTxn() BytesTxn
	// GetValue retrieves one item under the provided key
	GetValue(key string) (data []byte, found bool, revision int64, err error)
	// ListValues returns an iterator that enables to traverse all items stored under the provided key
	ListValues(key string) (BytesKeyValIterator, error)
	// ListKeys is similar to the ListValues the difference is that values are not fetched
	ListKeys(prefix string) (BytesKeyIterator, error)
	// Delete removes data stored under the key
	Delete(key string) (bool, error)
}

// BytesKeyVal represents a single key-value pair
type BytesKeyVal interface {
	// GetKey returns the key of the pair
	GetKey() string
	// GetValue returns the value of the pair
	GetValue() []byte
	// GetRevision returns revision associated with the latest change in the key-value pair
	GetRevision() int64
}

// BytesKeyValIterator is an iterator returned by ListValues call
type BytesKeyValIterator interface {
	// GetNext retrieves the following item from the context.
	GetNext() (kv BytesKeyVal, allReceived bool)
}

// BytesKeyIterator is an iterator returned by ListKeys call
type BytesKeyIterator interface {
	// GetNext retrieves the following item from the context.
	GetNext() (key string, rev int64, allReceived bool)
}

// BytesPluginBroker is a wrapper of BytesBroker. It allows to share BytesBroker's connection
// to datastore. Every BytesPluginBroker is able to create/modify/delete key-value
// pairs and watch distinct set of keys.
type BytesPluginBroker interface {
	BytesBroker
	BytesWatcher
}
