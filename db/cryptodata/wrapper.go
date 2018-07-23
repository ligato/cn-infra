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

package cryptodata

import (
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/datasync"
)

// CoreBrokerWatcherWrapper wraps keyval.CoreBrokerWatcher with additional support of reading encrypted data
type CoreBrokerWatcherWrapper struct {
	// Wrapped CoreBrokerWatcher
	wrap keyval.CoreBrokerWatcher
	// Wrapped BytesBroker
	bytesWrap *BytesBrokerWrapper
	// Function used for decrypting arbitrary data later
	decryptArbitrary DecryptArbitrary
	// Decrypter is used to decrypt data
	decrypter Decrypter
}

// BytesBrokerWrapper wraps keyval.BytesBroker with additional support of reading encrypted data
type BytesBrokerWrapper struct {
	// Wrapped CoreBrokerWatcher
	wrap keyval.BytesBroker
	// Function used for decrypting arbitrary data later
	decryptArbitrary DecryptArbitrary
	// Decrypter is used to decrypt data
	decrypter Decrypter
}

// NewCoreBrokerWatcherWrapper creates wrapper for provided CoreBrokerWatcher, adding support for decrypting encrypted
// data
func NewCoreBrokerWatcherWrapper(cbw keyval.CoreBrokerWatcher, decrypter Decrypter, decryptArbitrary DecryptArbitrary) *CoreBrokerWatcherWrapper {
	return &CoreBrokerWatcherWrapper{
		wrap:             cbw,
		decryptArbitrary: decryptArbitrary,
		decrypter:        decrypter,
		bytesWrap: &BytesBrokerWrapper{
			wrap:             cbw,
			decryptArbitrary: decryptArbitrary,
			decrypter:        decrypter,
		},
	}
}

// Watch starts subscription for changes associated with the selected keys.
// Watch events will be delivered to callback (not channel) <respChan>.
// Channel <closeChan> can be used to close watching on respective key
func (cbw *CoreBrokerWatcherWrapper) Watch(respChan func(keyval.BytesWatchResp), closeChan chan string, keys ...string) error {
	return cbw.wrap.Watch(respChan, closeChan, keys...)
}

// NewBroker returns a BytesBroker instance that prepends given
// <keyPrefix> to all keys in its calls.
// To avoid using a prefix, pass keyval.Root constant as argument.
func (cbw *CoreBrokerWatcherWrapper) NewBroker(prefix string) keyval.BytesBroker {
	return &BytesBrokerWrapper{
		wrap:             cbw.wrap.NewBroker(prefix),
		decryptArbitrary: cbw.decryptArbitrary,
		decrypter:        cbw.decrypter,
	}
}

// NewWatcher returns a BytesWatcher instance. Given <keyPrefix> is
// prepended to keys during watch subscribe phase.
// The prefix is removed from the key retrieved by GetKey() in BytesWatchResp.
// To avoid using a prefix, pass keyval.Root constant as argument.
func (cbw *CoreBrokerWatcherWrapper) NewWatcher(prefix string) keyval.BytesWatcher {
	return cbw.wrap.NewWatcher(prefix)
}

// Close closes provided wrapper
func (cbw *CoreBrokerWatcherWrapper) Close() error {
	return cbw.wrap.Close()
}

// Put puts single key-value pair into db.
// The behavior of put can be adjusted using PutOptions.
func (cbw *CoreBrokerWatcherWrapper) Put(key string, data []byte, opts ...datasync.PutOption) error {
	return cbw.bytesWrap.Put(key, data, opts...)
}

// NewTxn creates a transaction.
func (cbw *CoreBrokerWatcherWrapper) NewTxn() keyval.BytesTxn {
	return cbw.bytesWrap.NewTxn()
}

// GetValue retrieves one item under the provided key.
func (cbw *CoreBrokerWatcherWrapper) GetValue(key string) (data []byte, found bool, revision int64, err error) {
	return cbw.bytesWrap.GetValue(key)
}

// ListValues returns an iterator that enables to traverse all items stored
// under the provided <key>.
func (cbw *CoreBrokerWatcherWrapper) ListValues(key string) (keyval.BytesKeyValIterator, error) {
	return cbw.bytesWrap.ListValues(key)
}

// ListKeys returns an iterator that allows to traverse all keys from data
// store that share the given <prefix>.
func (cbw *CoreBrokerWatcherWrapper) ListKeys(prefix string) (keyval.BytesKeyIterator, error) {
	return cbw.bytesWrap.ListKeys(prefix)
}

// Delete removes data stored under the <key>.
func (cbw *CoreBrokerWatcherWrapper) Delete(key string, opts ...datasync.DelOption) (existed bool, err error) {
	return cbw.bytesWrap.Delete(key, opts...)
}

// Put puts single key-value pair into db.
// The behavior of put can be adjusted using PutOptions.
func (cbb *BytesBrokerWrapper) Put(key string, data []byte, opts ...datasync.PutOption) error {
	return cbb.wrap.Put(key, data, opts...)
}

// NewTxn creates a transaction.
func (cbb *BytesBrokerWrapper) NewTxn() keyval.BytesTxn {
	return cbb.wrap.NewTxn()
}

// GetValue retrieves one item under the provided key.
func (cbb *BytesBrokerWrapper) GetValue(key string) (data []byte, found bool, revision int64, err error) {
	data, found, revision, err = cbb.wrap.GetValue(key)
	if err == nil {
		data = cbb.decrypter.Decrypt(data, cbb.decryptArbitrary)
	}
	return
}

// ListValues returns an iterator that enables to traverse all items stored
// under the provided <key>.
func (cbb *BytesBrokerWrapper) ListValues(key string) (keyval.BytesKeyValIterator, error) {
	return cbb.wrap.ListValues(key)
}

// ListKeys returns an iterator that allows to traverse all keys from data
// store that share the given <prefix>.
func (cbb *BytesBrokerWrapper) ListKeys(prefix string) (keyval.BytesKeyIterator, error) {
	return cbb.wrap.ListKeys(prefix)
}

// Delete removes data stored under the <key>.
func (cbb *BytesBrokerWrapper) Delete(key string, opts ...datasync.DelOption) (existed bool, err error) {
	return cbb.wrap.Delete(key, opts...)
}
