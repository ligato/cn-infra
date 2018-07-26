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
	"github.com/golang/protobuf/proto"
)

// KvBytesPluginWrapper wraps keyval.KvBytesPlugin with additional support of reading encrypted data
type KvBytesPluginWrapper struct {
	// Wrapped KvBytesPlugin
	keyval.KvBytesPlugin
	// Function used for decrypting arbitrary data later
	decryptFunc DecryptFunc
	// ArbitraryDecrypter is used to decrypt data
	decrypter ArbitraryDecrypter
}

// BytesBrokerWrapper wraps keyval.BytesBroker with additional support of reading encrypted data
type BytesBrokerWrapper struct {
	// Wrapped BytesBroker
	keyval.BytesBroker
	// Function used for decrypting arbitrary data later
	decryptFunc DecryptFunc
	// ArbitraryDecrypter is used to decrypt data
	decrypter ArbitraryDecrypter
}

// NewKvBytesPluginWrapper creates wrapper for provided CoreBrokerWatcher, adding support for decrypting encrypted
// data
func NewKvBytesPluginWrapper(cbw keyval.KvBytesPlugin, decrypter ArbitraryDecrypter, decryptFunc DecryptFunc) *KvBytesPluginWrapper {
	return &KvBytesPluginWrapper{
		KvBytesPlugin: cbw,
		decryptFunc:   decryptFunc,
		decrypter:     decrypter,
	}
}

// NewBroker returns a BytesBroker instance with support for decrypting values that prepends given <keyPrefix> to all
// keys in its calls.
// To avoid using a prefix, pass keyval.Root constant as argument.
func (cbw *KvBytesPluginWrapper) NewBroker(prefix string) keyval.BytesBroker {
	return NewBytesBrokerWrapper(cbw.KvBytesPlugin.NewBroker(prefix), cbw.decrypter, cbw.decryptFunc)
}

// NewBytesBrokerWrapper creates wrapper for provided BytesBroker, adding support for decrypting encrypted data
func NewBytesBrokerWrapper(pb keyval.BytesBroker, decrypter ArbitraryDecrypter, decryptFunc DecryptFunc) *BytesBrokerWrapper {
	return &BytesBrokerWrapper{
		BytesBroker: pb,
		decryptFunc: decryptFunc,
		decrypter:   decrypter,
	}
}

// GetValue retrieves and tries to decrypt one item under the provided key.
func (cbb *BytesBrokerWrapper) GetValue(key string) (data []byte, found bool, revision int64, err error) {
	data, found, revision, err = cbb.BytesBroker.GetValue(key)
	if err == nil {
		objData, err := cbb.decrypter.Decrypt(data, cbb.decryptFunc)
		if err != nil {
			return data, found, revision, err
		}

		data = objData.([]byte)
	}
	return
}

// KvProtoPluginWrapper wraps keyval.KvProtoPlugin with additional support of reading encrypted data
type KvProtoPluginWrapper struct {
	// Wrapped KvProtoPlugin
	keyval.KvProtoPlugin
	// Function used for decrypting arbitrary data later
	decryptFunc DecryptFunc
	// ArbitraryDecrypter is used to decrypt data
	decrypter ArbitraryDecrypter
}

// ProtoBrokerWrapper wraps keyval.ProtoBroker with additional support of reading encrypted data
type ProtoBrokerWrapper struct {
	keyval.ProtoBroker
	// Function used for decrypting arbitrary data later
	decryptFunc DecryptFunc
	// ArbitraryDecrypter is used to decrypt data
	decrypter ArbitraryDecrypter
}

// NewKvProtoPluginWrapper creates wrapper for provided KvProtoPlugin, adding support for decrypting encrypted data
func NewKvProtoPluginWrapper(kvp keyval.KvProtoPlugin, decrypter ArbitraryDecrypter, decryptFunc DecryptFunc) *KvProtoPluginWrapper {
	return &KvProtoPluginWrapper{
		KvProtoPlugin: kvp,
		decryptFunc:   decryptFunc,
		decrypter:     decrypter,
	}
}

// NewBroker returns a BytesBroker instance with support for decrypting values that prepends given <keyPrefix> to all
// keys in its calls.
// To avoid using a prefix, pass keyval.Root constant as argument.
func (kvp *KvProtoPluginWrapper) NewBroker(prefix string) keyval.ProtoBroker {
	return NewProtoBrokerWrapper(kvp.KvProtoPlugin.NewBroker(prefix), kvp.decrypter, kvp.decryptFunc)
}

// NewProtoBrokerWrapper creates wrapper for provided ProtoBroker, adding support for decrypting encrypted data
func NewProtoBrokerWrapper(pb keyval.ProtoBroker, decrypter ArbitraryDecrypter, decryptFunc DecryptFunc) *ProtoBrokerWrapper {
	return &ProtoBrokerWrapper{
		ProtoBroker: pb,
		decryptFunc: decryptFunc,
		decrypter:   decrypter,
	}
}

// GetValue retrieves one item under the provided <key>. If the item exists,
// it is unmarshaled into the <reqObj> and its fields are decrypted.
func (db *ProtoBrokerWrapper) GetValue(key string, reqObj proto.Message) (bool, int64, error) {
	found, revision, err := db.ProtoBroker.GetValue(key, reqObj)
	if !found || err != nil {
		return found, revision, err
	}

	_, err = db.decrypter.Decrypt(reqObj, db.decryptFunc)
	return found, revision, err
}
