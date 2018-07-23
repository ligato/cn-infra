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
)

// CoreBrokerWatcherWrapper wraps keyval.CoreBrokerWatcher with additional support of reading encrypted data
type CoreBrokerWatcherWrapper struct {
	// Wrapped CoreBrokerWatcher
	keyval.CoreBrokerWatcher
	// Wrapped BytesBroker
	bytesWrap *BytesBrokerWrapper
	// Function used for decrypting arbitrary data later
	decryptArbitrary DecryptArbitrary
	// Decrypter is used to decrypt data
	decrypter Decrypter
}

// BytesBrokerWrapper wraps keyval.BytesBroker with additional support of reading encrypted data
type BytesBrokerWrapper struct {
	// Wrapped BytesBroker
	keyval.BytesBroker
	// Function used for decrypting arbitrary data later
	decryptArbitrary DecryptArbitrary
	// Decrypter is used to decrypt data
	decrypter Decrypter
}

// NewCoreBrokerWatcherWrapper creates wrapper for provided CoreBrokerWatcher, adding support for decrypting encrypted
// data
func NewCoreBrokerWatcherWrapper(cbw keyval.CoreBrokerWatcher, decrypter Decrypter, decryptArbitrary DecryptArbitrary) *CoreBrokerWatcherWrapper {
	return &CoreBrokerWatcherWrapper{
		CoreBrokerWatcher: cbw,
		decryptArbitrary:  decryptArbitrary,
		decrypter:         decrypter,
		bytesWrap: &BytesBrokerWrapper{
			BytesBroker:      cbw,
			decryptArbitrary: decryptArbitrary,
			decrypter:        decrypter,
		},
	}
}

// NewBroker returns a BytesBroker instance with support for decrypting values that prepends given <keyPrefix> to all
// keys in its calls.
// To avoid using a prefix, pass keyval.Root constant as argument.
func (cbw *CoreBrokerWatcherWrapper) NewBroker(prefix string) keyval.BytesBroker {
	return &BytesBrokerWrapper{
		BytesBroker:      cbw.CoreBrokerWatcher.NewBroker(prefix),
		decryptArbitrary: cbw.decryptArbitrary,
		decrypter:        cbw.decrypter,
	}
}

// GetValue retrieves and tries to decrypt one item under the provided key.
func (cbb *BytesBrokerWrapper) GetValue(key string) (data []byte, found bool, revision int64, err error) {
	data, found, revision, err = cbb.BytesBroker.GetValue(key)
	if err == nil {
		data = cbb.decrypter.Decrypt(data, cbb.decryptArbitrary)
	}
	return
}
