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
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/db/keyval"
)

// ProtoWrapperEtcd is a decorator which allows to read/write proto file modelled data.
// It marshals/unmarshals go structures to slice off bytes and vice versa behind the scenes.
type ProtoWrapperEtcd struct {
	broker     *BytesConnectionEtcd
	serializer keyval.Serializer
}

type protoBrokerEtcd struct {
	broker     keyval.BytesBroker
	serializer keyval.Serializer
}

type protoWatcherEtcd struct {
	watcher    keyval.BytesWatcher
	serializer keyval.Serializer
}

// protoKeyValIterator is an iterator returned by ListValues call
type protoKeyValIterator struct {
	ctx        keyval.BytesKeyValIterator
	serializer keyval.Serializer
}

// protoKeyIterator is an iterator returned by ListKeys call
type protoKeyIterator struct {
	ctx keyval.BytesKeyIterator
}

// protoKeyVal represents single key-value pair
type protoKeyVal struct {
	pair       keyval.BytesKeyVal
	serializer keyval.Serializer
}

// NewProtoWrapperEtcd initializes proto decorator. The default serializer is used - SerializerProto.
func NewProtoWrapperEtcd(db *BytesConnectionEtcd) *ProtoWrapperEtcd {
	return &ProtoWrapperEtcd{db, &keyval.SerializerProto{}}
}

// NewProtoWrapperWithSerializer initializes proto decorator with the specified serializer.
func NewProtoWrapperWithSerializer(db *BytesConnectionEtcd, serializer keyval.Serializer) *ProtoWrapperEtcd {
	return &ProtoWrapperEtcd{db, serializer}
}

// Close closes underlying connection to etcd. Beware: if the connection is shared among multiple instances this might
// unintentionally cancel the connection.
func (db *ProtoWrapperEtcd) Close() error {
	return db.broker.Close()
}

// NewBroker creates a new instance of the proxy that shares the underlying connection
// and allows to read/edit key-value pairs.
func (db *ProtoWrapperEtcd) NewBroker(prefix string) keyval.ProtoBroker {
	return &protoBrokerEtcd{db.broker.NewBroker(prefix), db.serializer}
}

// NewWatcher creates a new instance of the proxy that shares the underlying connection
// and allows to subscribe for watching of the changes.
func (db *ProtoWrapperEtcd) NewWatcher(prefix string) keyval.ProtoWatcher {
	return &protoWatcherEtcd{db.broker.NewWatcher(prefix), db.serializer}
}

// NewTxn creates a new Data Broker transaction. A transaction can
// holds multiple operations that are all committed to the data
// store together. After a transaction has been created, one or
// more operations (put or delete) can be added to the transaction
// before it is committed.
func (db *ProtoWrapperEtcd) NewTxn() keyval.ProtoTxn {
	return &protoTxn{txn: db.broker.NewTxn(), serializer: db.serializer}
}

// NewTxn creates a new Data Broker transaction. A transaction can
// hold multiple operations that are all committed to the data
// store together. After a transaction has been created, one or
// more operations (put or delete) can be added to the transaction
// before it is committed.
func (pdb *protoBrokerEtcd) NewTxn() keyval.ProtoTxn {
	return &protoTxn{txn: pdb.broker.NewTxn(), serializer: pdb.serializer}
}

// Put writes the provided key-value item into the data store.
//
// Returns an error if the item could not be written, ok otherwise.
func (db *ProtoWrapperEtcd) Put(key string, value proto.Message, opts ...keyval.PutOption) error {
	return putProtoInternal(db.broker, db.serializer, key, value, opts...)
}

// Put writes the provided key-value item into the data store.
//
// Returns an error if the item could not be written, ok otherwise.
func (pdb *protoBrokerEtcd) Put(key string, value proto.Message, opts ...keyval.PutOption) error {
	return putProtoInternal(pdb.broker, pdb.serializer, key, value, opts...)
}

func putProtoInternal(broker keyval.BytesBroker, serializer keyval.Serializer, key string, value proto.Message, opts ...keyval.PutOption) error {
	// Marshal value to protobuf
	binData, err := serializer.Marshal(value)
	if err != nil {
		return err
	}
	broker.Put(key, binData, opts...)
	return nil
}

// Delete removes from datastore key-value items stored under key.
func (db *ProtoWrapperEtcd) Delete(key string) (existed bool, err error) {
	return db.broker.Delete(key)
}

// Delete removes from datastore key-value items stored under key.
func (pdb *protoBrokerEtcd) Delete(key string) (existed bool, err error) {
	return pdb.broker.Delete(key)
}

// Watch subscribes for changes in datastore associated with the key. respChannel is used for delivery watch events
func (db *ProtoWrapperEtcd) Watch(resp chan keyval.ProtoWatchResp, keys ...string) error {
	byteCh := make(chan keyval.BytesWatchResp, 0)
	err := db.broker.Watch(byteCh, keys...)
	if err != nil {
		return err
	}
	go func() {
		for msg := range byteCh {
			resp <- NewWatchResp(db.serializer, msg)
		}
	}()
	return nil
}

// GetValue retrieves one key-value item from the datastore. The item
// is identified by the provided key.
//
// If the item was found, its value is unmarshaled and placed in
// the `reqObj` message buffer and the function returns found=true.
// If the object was not found, the function returns found=false.
// Function returns revision=revision of the latest modification
// If an error was encountered, the function returns an error.
func (db *ProtoWrapperEtcd) GetValue(key string, reqObj proto.Message) (found bool, revision int64, err error) {
	return getValueProtoInternal(db.broker, db.serializer, key, reqObj)
}

// GetValue retrieves one key-value item from the datastore. The item
// is identified by the provided key.
//
// If the item was found, its value is unmarshaled and placed in
// the `reqObj` message buffer and the function returns found=true.
// If the object was not found, the function returns found=false.
// Function returns revision=revision of the latest modification
// If an error was encountered, the function returns an error.
func (pdb *protoBrokerEtcd) GetValue(key string, reqObj proto.Message) (found bool, revision int64, err error) {
	return getValueProtoInternal(pdb.broker, pdb.serializer, key, reqObj)
}

func getValueProtoInternal(broker keyval.BytesBroker, serializer keyval.Serializer, key string, reqObj proto.Message) (found bool, revision int64, err error) {
	// get data from etcdv3
	resp, found, rev, err := broker.GetValue(key)
	if err != nil {
		return false, 0, err
	}

	if !found {
		return false, 0, nil
	}

	err = serializer.Unmarshal(resp, reqObj)
	if err != nil {
		return false, 0, err
	}
	return true, rev, nil
}

// ListValues retrieves an iterator for elements stored under the provided key.
func (db *ProtoWrapperEtcd) ListValues(key string) (keyval.ProtoKeyValIterator, error) {
	return listValuesProtoInternal(db.broker, db.serializer, key)
}

// ListValues retrieves an iterator for elements stored under the provided key.
func (pdb *protoBrokerEtcd) ListValues(key string) (keyval.ProtoKeyValIterator, error) {
	return listValuesProtoInternal(pdb.broker, pdb.serializer, key)
}

func listValuesProtoInternal(broker keyval.BytesBroker, serializer keyval.Serializer, key string) (keyval.ProtoKeyValIterator, error) {
	// get data from etcdv3
	ctx, err := broker.ListValues(key)
	if err != nil {
		return nil, err
	}
	return &protoKeyValIterator{ctx, serializer}, nil
}

// ListKeys is similar to the ListValues the difference is that values are not fetched
func (db *ProtoWrapperEtcd) ListKeys(prefix string) (keyval.ProtoKeyIterator, error) {
	return listKeysProtoInternal(db.broker, prefix)
}

// ListKeys is similar to the ListValues the difference is that values are not fetched
func (pdb *protoBrokerEtcd) ListKeys(prefix string) (keyval.ProtoKeyIterator, error) {
	return listKeysProtoInternal(pdb.broker, prefix)
}

func listKeysProtoInternal(broker keyval.BytesBroker, prefix string) (keyval.ProtoKeyIterator, error) {
	// get data from etcdv3
	ctx, err := broker.ListKeys(prefix)
	if err != nil {
		return nil, err
	}
	return &protoKeyIterator{ctx}, nil
}

// ListValuesRange retrieves an iterator for elements stored in specified range.
func (db *ProtoWrapperEtcd) ListValuesRange(fromPrefix string, toPrefix string) (keyval.ProtoKeyValIterator, error) {

	ctx, err := db.broker.ListValuesRange(fromPrefix, toPrefix)
	if err != nil {
		return nil, err
	}
	return &protoKeyValIterator{ctx, db.serializer}, nil
}

// GetNext returns the following item from the result set. If data was returned, found is set to true.
func (ctx *protoKeyValIterator) GetNext() (kv keyval.ProtoKeyVal, lastReceived bool) {
	pair, allReceived := ctx.ctx.GetNext()
	if allReceived {
		return nil, allReceived
	}

	return &protoKeyVal{pair, ctx.serializer}, allReceived
}

// GetNext returns the following item from the result set. If data was returned, found is set to true.
func (ctx *protoKeyIterator) GetNext() (key string, rev int64, lastReceived bool) {
	return ctx.ctx.GetNext()
}

// Watch for changes in datastore respChannel is used for receiving watch events
func (pdb *protoWatcherEtcd) Watch(resp chan keyval.ProtoWatchResp, keys ...string) error {
	byteCh := make(chan keyval.BytesWatchResp, 0)
	err := pdb.watcher.Watch(byteCh, keys...)
	if err != nil {
		return err
	}
	go func() {
		for msg := range byteCh {
			resp <- NewWatchResp(pdb.serializer, msg)
		}
	}()
	return nil
}

// GetValue returns the value of the pair
func (kv *protoKeyVal) GetValue(msg proto.Message) error {
	err := kv.serializer.Unmarshal(kv.pair.GetValue(), msg)
	if err != nil {
		return err
	}
	return nil
}

// GetKey returns the key of the pair
func (kv *protoKeyVal) GetKey() string {
	return kv.pair.GetKey()
}

// GetRevision returns the revision associated with the pair
func (kv *protoKeyVal) GetRevision() int64 {
	return kv.pair.GetRevision()
}
