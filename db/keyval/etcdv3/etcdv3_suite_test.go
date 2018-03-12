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
	"errors"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/logging/logrus"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
)

var dataBroker *BytesConnectionEtcd
var dataBrokerErr *BytesConnectionEtcd
var pluginDataBroker *BytesBrokerWatcherEtcd

// Mock data broker err
type MockKVErr struct {
	// NO-OP
}

func (mock *MockKVErr) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return nil, errors.New("test-error")
}

func (mock *MockKVErr) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return nil, errors.New("test-error")
}

func (mock *MockKVErr) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return nil, errors.New("test-error")
}

func (mock *MockKVErr) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	return nil, nil
}

func (mock *MockKVErr) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	return clientv3.OpResponse{}, nil
}

func (mock *MockKVErr) Txn(ctx context.Context) clientv3.Txn {
	return &MockTxn{}
}

func (mock *MockKVErr) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return nil
}

func (mock *MockKVErr) Close() error {
	return nil
}

// Mock KV
type MockKV struct {
	// NO-OP
}

func (mock *MockKV) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return nil, nil
}

func (mock *MockKV) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	response := *new(clientv3.GetResponse)
	kvs := new(mvccpb.KeyValue)
	kvs.Key = []byte{1}
	kvs.Value = []byte{73, 0x6f, 0x6d, 65, 0x2d, 0x6a, 73, 0x6f, 0x6e} //some-json
	response.Kvs = []*mvccpb.KeyValue{kvs}
	return &response, nil
}

func (mock *MockKV) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	response := *new(clientv3.DeleteResponse)
	response.PrevKvs = []*mvccpb.KeyValue{}
	return &response, nil
}

func (mock *MockKV) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	return nil, nil
}

func (mock *MockKV) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	return clientv3.OpResponse{}, nil
}

func (mock *MockKV) Txn(ctx context.Context) clientv3.Txn {
	return &MockTxn{}
}

func (mock *MockKV) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return nil
}

func (mock *MockKV) Close() error {
	return nil
}

// Mock Txn
type MockTxn struct {
}

func (mock *MockTxn) If(cs ...clientv3.Cmp) clientv3.Txn {
	return &MockTxn{}
}

func (mock *MockTxn) Then(ops ...clientv3.Op) clientv3.Txn {
	return &MockTxn{}
}

func (mock *MockTxn) Else(ops ...clientv3.Op) clientv3.Txn {
	return &MockTxn{}
}

func (mock *MockTxn) Commit() (*clientv3.TxnResponse, error) {
	return nil, nil
}

// Tests

func init() {
	mockKv := &MockKV{}
	mockKvErr := &MockKVErr{}
	dataBroker = &BytesConnectionEtcd{Logger: logrus.DefaultLogger(), etcdClient: &clientv3.Client{KV: mockKv, Watcher: mockKv}}
	dataBrokerErr = &BytesConnectionEtcd{Logger: logrus.DefaultLogger(), etcdClient: &clientv3.Client{KV: mockKvErr, Watcher: mockKvErr}}
	pluginDataBroker = &BytesBrokerWatcherEtcd{Logger: logrus.DefaultLogger(), closeCh: make(chan string), kv: mockKv, watcher: mockKv}
}

func TestNewTxn(t *testing.T) {
	RegisterTestingT(t)
	newTxn := dataBroker.NewTxn()
	Expect(newTxn).NotTo(BeNil())
}

func TestTxnPut(t *testing.T) {
	RegisterTestingT(t)
	newTxn := dataBroker.NewTxn()
	result := newTxn.Put("key", []byte("data"))
	Expect(result).NotTo(BeNil())
}

func TestTxnDelete(t *testing.T) {
	RegisterTestingT(t)
	newTxn := dataBroker.NewTxn()
	Expect(newTxn).NotTo(BeNil())
	result := newTxn.Delete("key")
	Expect(result).NotTo(BeNil())
}

func TestTxnCommit(t *testing.T) {
	RegisterTestingT(t)
	newTxn := dataBroker.NewTxn()
	result := newTxn.Commit()
	Expect(result).To(BeNil())
}

func TestPut(t *testing.T) {
	// regular case
	RegisterTestingT(t)
	err := dataBroker.Put("key", []byte("data"))
	Expect(err).ShouldNot(HaveOccurred())
	// error case
	err = dataBrokerErr.Put("key", []byte("data"))
	Expect(err).Should(HaveOccurred())
	Expect(err.Error()).To(BeEquivalentTo("test-error"))
}

func TestGetValue(t *testing.T) {
	// regular case
	RegisterTestingT(t)
	result, found, _, err := dataBroker.GetValue("key")
	Expect(err).ShouldNot(HaveOccurred())
	Expect(result).NotTo(BeNil())
	// error case
	result, found, _, err = dataBrokerErr.GetValue("key")
	Expect(err).Should(HaveOccurred())
	Expect(found).To(BeFalse())
	Expect(result).To(BeNil())
	Expect(err.Error()).To(BeEquivalentTo("test-error"))
}

func TestListValues(t *testing.T) {
	// regular case
	RegisterTestingT(t)
	result, err := dataBroker.ListValues("key")
	Expect(err).ShouldNot(HaveOccurred())
	Expect(result).ToNot(BeNil())

	// error case
	result, err = dataBrokerErr.ListValues("key")
	Expect(err).Should(HaveOccurred())
	Expect(result).To(BeNil())
	Expect(err.Error()).To(BeEquivalentTo("test-error"))
}

func TestListValuesRange(t *testing.T) {
	// regular case
	RegisterTestingT(t)
	result, err := dataBroker.ListValuesRange("AKey", "ZKey")
	Expect(err).ShouldNot(HaveOccurred())
	Expect(result).ToNot(BeNil())

	// error case
	result, err = dataBrokerErr.ListValuesRange("AKey", "ZKey")
	Expect(err).Should(HaveOccurred())
	Expect(result).To(BeNil())
	Expect(err.Error()).To(BeEquivalentTo("test-error"))
}

func TestDelete(t *testing.T) {
	// regular case
	RegisterTestingT(t)
	response, err := dataBroker.Delete("vnf")
	Expect(err).ShouldNot(HaveOccurred())
	Expect(response).To(BeFalse())
	// error case
	response, err = dataBrokerErr.Delete("vnf")
	Expect(err).Should(HaveOccurred())
	Expect(response).To(BeFalse())
	Expect(err.Error()).To(BeEquivalentTo("test-error"))
}

func TestNewBroker(t *testing.T) {
	RegisterTestingT(t)
	pdb := dataBroker.NewBroker("/pluginname")
	Expect(pdb).NotTo(BeNil())
}

func TestNewWatcher(t *testing.T) {
	RegisterTestingT(t)
	pdb := dataBroker.NewWatcher("/pluginname")
	Expect(pdb).NotTo(BeNil())
}

func TestWatch(t *testing.T) {
	RegisterTestingT(t)
	err := pluginDataBroker.Watch(func(keyval.BytesWatchResp) {}, nil, "key")
	Expect(err).ShouldNot(HaveOccurred())
}

func TestWatchPutResp(t *testing.T) {
	var rev int64 = 1
	value := []byte("data")
	prevVal := []byte("prevData")
	key := "key"
	RegisterTestingT(t)
	createResp := NewBytesWatchPutResp(key, value, prevVal, rev)
	Expect(createResp).NotTo(BeNil())
	Expect(createResp.GetChangeType()).To(BeEquivalentTo(datasync.Put))
	Expect(createResp.GetKey()).To(BeEquivalentTo(key))
	Expect(createResp.GetValue()).To(BeEquivalentTo(value))
	Expect(createResp.GetPrevValue()).To(BeEquivalentTo(prevVal))
	Expect(createResp.GetRevision()).To(BeEquivalentTo(rev))
}

func TestWatchDeleteResp(t *testing.T) {
	var rev int64 = 1
	key := "key"
	prevVal := []byte("prevVal")
	RegisterTestingT(t)
	createResp := NewBytesWatchDelResp(key, prevVal, rev)
	Expect(createResp).NotTo(BeNil())
	Expect(createResp.GetChangeType()).To(BeEquivalentTo(datasync.Delete))
	Expect(createResp.GetKey()).To(BeEquivalentTo(key))
	Expect(createResp.GetValue()).To(BeNil())
	Expect(createResp.GetPrevValue()).To(BeEquivalentTo(prevVal))
	Expect(createResp.GetRevision()).To(BeEquivalentTo(rev))
}

func TestConfig(t *testing.T) {
	RegisterTestingT(t)
	cfg := &Config{DialTimeout: time.Second, OpTimeout: time.Second}
	etcdCfg, err := ConfigToClientv3(cfg)
	Expect(err).To(BeNil())
	Expect(etcdCfg).NotTo(BeNil())
	Expect(etcdCfg.OpTimeout).To(BeEquivalentTo(time.Second))
	Expect(etcdCfg.DialTimeout).To(BeEquivalentTo(time.Second))
	Expect(etcdCfg.TLS).To(BeNil())
}
