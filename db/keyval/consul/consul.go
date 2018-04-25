//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package consul

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/logging/logrus"

	"github.com/hashicorp/consul/api"
)

func transformKey(key string) string {
	return strings.TrimPrefix(key, "/")
}

// Store serves as a client for Consul KV storage and implements keyval.CoreBrokerWatcher interface.
type Store struct {
	client *api.Client
}

// NewConsulStore creates new client for Consul using given address.
func NewConsulStore(addr string) (store *Store, err error) {
	cfg := api.DefaultConfig()
	cfg.Address = addr

	var c *api.Client
	if c, err = api.NewClient(cfg); err != nil {
		return nil, fmt.Errorf("failed to create Consul client %s", err)
	}

	peers, err := c.Status().Peers()
	if err != nil {
		return nil, err
	}
	logrus.DefaultLogger().Debugf("consul peers: %v", peers)

	return &Store{
		client: c,
	}, nil

}

// Put stores given data for the key.
func (c *Store) Put(key string, data []byte, opts ...datasync.PutOption) error {
	fmt.Printf("put: %q\n", key)
	p := &api.KVPair{Key: transformKey(key), Value: data}
	_, err := c.client.KV().Put(p, nil)
	if err != nil {
		return err
	}

	return nil
}

// NewTxn creates new transaction.
func (c *Store) NewTxn() keyval.BytesTxn {
	panic("implement me")
}

// GetValue returns data for the given key.
func (c *Store) GetValue(key string) (data []byte, found bool, revision int64, err error) {
	fmt.Printf("get value: %q\n", key)
	pair, _, err := c.client.KV().Get(transformKey(key), nil)
	if err != nil {
		return nil, false, 0, err
	} else if pair == nil {
		return nil, false, 0, nil
	}

	return pair.Value, true, int64(pair.ModifyIndex), nil
}

// ListValues returns interator with key-value pairs for given key prefix.
func (c *Store) ListValues(key string) (keyval.BytesKeyValIterator, error) {
	pairs, _, err := c.client.KV().List(transformKey(key), nil)
	if err != nil {
		return nil, err
	}

	return &bytesKeyValIterator{len: len(pairs), pairs: pairs}, nil
}

// ListKeys returns interator with keys for given key prefix.
func (c *Store) ListKeys(prefix string) (keyval.BytesKeyIterator, error) {
	keys, _, err := c.client.KV().Keys(transformKey(prefix), "", nil)
	if err != nil {
		return nil, err
	}

	return &bytesKeyIterator{len: len(keys), keys: keys}, nil
}

// Delete deletes given key.
func (c *Store) Delete(key string, opts ...datasync.DelOption) (existed bool, err error) {
	fmt.Printf("delete: %q\n", key)
	if _, err := c.client.KV().Delete(transformKey(key), nil); err != nil {
		return false, err
	}

	return true, nil
}

// Watch watches given list of key prefixes.
func (c *Store) Watch(resp func(keyval.BytesWatchResp), closeChan chan string, keys ...string) error {
	for _, k := range keys {
		if err := c.watch(resp, closeChan, k); err != nil {
			return err
		}
	}
	return nil
}

func (c *Store) watch(resp func(watchResp keyval.BytesWatchResp), closeCh chan string, prefix string) error {
	logrus.DefaultLogger().Debug("WATCH:", prefix)

	ctx, cancel := context.WithCancel(context.Background())

	recvChan := c.watchPrefix(ctx, prefix)

	go func(regPrefix string) {
		for {
			select {
			case wr, ok := <-recvChan:
				if !ok {
					logrus.DefaultLogger().WithField("prefix", prefix).Debug("Watch recv chan was closed")
					return
				}
				for _, ev := range wr.Events {
					var r keyval.BytesWatchResp
					if ev.Type == datasync.Put {
						r = etcdv3.NewBytesWatchPutResp(ev.Key, ev.Value, ev.PrevValue, ev.Revision)
					} else {
						r = etcdv3.NewBytesWatchDelResp(ev.Key, ev.Value, ev.Revision)
					}
					resp(r)
				}
			case closeVal, ok := <-closeCh:
				if !ok || closeVal == regPrefix {
					logrus.DefaultLogger().WithField("prefix", prefix).Debug("Watch ended")
					cancel()
					return
				}
			}
		}
	}(prefix)

	return nil
}

type watchEvent struct {
	Type      datasync.PutDel
	Key       string
	Value     []byte
	PrevValue []byte
	Revision  int64
}

type watchResponse struct {
	Events []*watchEvent
	Err    error
}

func (c *Store) watchPrefix(ctx context.Context, prefix string) <-chan watchResponse {
	logrus.DefaultLogger().Debug("watchPrefix:", prefix)

	ch := make(chan watchResponse, 1)

	// Retrieve KV pairs and latest index
	qOpt := &api.QueryOptions{}
	oldPairs, qm, err := c.client.KV().List(prefix, qOpt.WithContext(ctx))
	if err != nil {
		ch <- watchResponse{Err: err}
		close(ch)
		return ch
	}

	oldIndex := qm.LastIndex
	oldPairsMap := make(map[string]*api.KVPair)

	logrus.DefaultLogger().Debugf("..retrieved: %v old pairs (old index: %v)\n", len(oldPairs), oldIndex)
	for _, pair := range oldPairs {
		logrus.DefaultLogger().Debugf(" - key: %q create: %v modify: %v\n", pair.Key, pair.CreateIndex, pair.ModifyIndex)
		oldPairsMap[pair.Key] = pair
	}

	go func() {
		for {
			logrus.DefaultLogger().Debug("calling list with wait")

			// Wait for an update to occur since the last index
			var newPairs api.KVPairs
			qOpt := &api.QueryOptions{
				WaitIndex: oldIndex,
			}
			newPairs, qm, err = c.client.KV().List(prefix, qOpt.WithContext(ctx))
			if err != nil {
				ch <- watchResponse{Err: err}
				close(ch)
				return
			}
			newIndex := qm.LastIndex

			// If the index is same as old one, request probably timed out, so we start again
			if oldIndex == newIndex {
				logrus.DefaultLogger().Debug("index unchanged, next round")
				continue
			}

			logrus.DefaultLogger().Debugf("..waited: %v new pairs (new index: %v) %+v\n", len(newPairs), newIndex, qm)
			for _, pair := range newPairs {
				logrus.DefaultLogger().Debugf(" + key: %q create: %v modify: %v\n", pair.Key, pair.CreateIndex, pair.ModifyIndex)
			}

			var evs []*watchEvent

			// Search for all created and modified KV
			for _, pair := range newPairs {
				if pair.ModifyIndex > oldIndex {
					var prevVal []byte
					if oldPair, ok := oldPairsMap[pair.Key]; ok {
						prevVal = oldPair.Value
					}
					evs = append(evs, &watchEvent{
						Type:      datasync.Put,
						Key:       pair.Key,
						Value:     pair.Value,
						PrevValue: prevVal,
						Revision:  int64(pair.ModifyIndex),
					})
				}
				delete(oldPairsMap, pair.Key)
			}
			// Search for all deleted KV
			for _, pair := range oldPairsMap {
				evs = append(evs, &watchEvent{
					Type:      datasync.Delete,
					Key:       pair.Key,
					PrevValue: pair.Value,
					Revision:  int64(pair.ModifyIndex),
				})
			}

			// Prepare latest KV pairs and last index for next round
			oldIndex = newIndex
			oldPairsMap = make(map[string]*api.KVPair)
			for _, pair := range newPairs {
				oldPairsMap[pair.Key] = pair
			}

			ch <- watchResponse{Events: evs}
		}
	}()
	return ch
}

// Close returns nil.
func (c *Store) Close() error {
	return nil
}

// NewBroker creates a new instance of a proxy that provides
// access to etcd. The proxy will reuse the connection from Store.
// <prefix> will be prepended to the key argument in all calls from the created
// BrokerWatcher. To avoid using a prefix, pass keyval. Root constant as
// an argument.
func (c *Store) NewBroker(prefix string) keyval.BytesBroker {
	return &BrokerWatcher{
		Store:  c,
		prefix: prefix,
	}
}

// NewWatcher creates a new instance of a proxy that provides
// access to etcd. The proxy will reuse the connection from Store.
// <prefix> will be prepended to the key argument in all calls on created
// BrokerWatcher. To avoid using a prefix, pass keyval. Root constant as
// an argument.
func (c *Store) NewWatcher(prefix string) keyval.BytesWatcher {
	return &BrokerWatcher{
		Store:  c,
		prefix: prefix,
	}
}

// BrokerWatcher uses Store to access the datastore.
// The connection can be shared among multiple BrokerWatcher.
// In case of accessing a particular subtree in Consul only,
// BrokerWatcher allows defining a keyPrefix that is prepended
// to all keys in its methods in order to shorten keys used in arguments.
type BrokerWatcher struct {
	*Store
	prefix string
}

func (pdb *BrokerWatcher) prefixKey(key string) string {
	return filepath.Join(pdb.prefix, key)
}

// Put calls 'Put' function of the underlying BytesConnectionEtcd.
// KeyPrefix defined in constructor is prepended to the key argument.
func (pdb *BrokerWatcher) Put(key string, data []byte, opts ...datasync.PutOption) error {
	return pdb.Store.Put(pdb.prefixKey(key), data, opts...)
}

// NewTxn creates a new transaction.
// KeyPrefix defined in constructor will be prepended to all key arguments
// in the transaction.
func (pdb *BrokerWatcher) NewTxn() keyval.BytesTxn {
	return pdb.Store.NewTxn()
}

// GetValue calls 'GetValue' function of the underlying BytesConnectionEtcd.
// KeyPrefix defined in constructor is prepended to the key argument.
func (pdb *BrokerWatcher) GetValue(key string) (data []byte, found bool, revision int64, err error) {
	return pdb.Store.GetValue(pdb.prefixKey(key))
}

// ListValues calls 'ListValues' function of the underlying BytesConnectionEtcd.
// KeyPrefix defined in constructor is prepended to the key argument.
// The prefix is removed from the keys of the returned values.
func (pdb *BrokerWatcher) ListValues(key string) (keyval.BytesKeyValIterator, error) {
	return pdb.Store.ListValues(pdb.prefixKey(key))
}

// ListKeys calls 'ListKeys' function of the underlying BytesConnectionEtcd.
// KeyPrefix defined in constructor is prepended to the argument.
func (pdb *BrokerWatcher) ListKeys(prefix string) (keyval.BytesKeyIterator, error) {
	return pdb.Store.ListKeys(pdb.prefixKey(prefix))
}

// Delete calls 'Delete' function of the underlying BytesConnectionEtcd.
// KeyPrefix defined in constructor is prepended to the key argument.
func (pdb *BrokerWatcher) Delete(key string, opts ...datasync.DelOption) (existed bool, err error) {
	return pdb.Store.Delete(pdb.prefixKey(key), opts...)
}

// Watch starts subscription for changes associated with the selected <keys>.
// KeyPrefix defined in constructor is prepended to all <keys> in the argument
// list. The prefix is removed from the keys returned in watch events.
// Watch events will be delivered to <resp> callback.
func (pdb *BrokerWatcher) Watch(resp func(keyval.BytesWatchResp), closeChan chan string, keys ...string) error {
	var prefixedKeys []string
	for _, key := range keys {
		prefixedKeys = append(prefixedKeys, pdb.prefixKey(key))
	}
	return pdb.Store.Watch(resp, closeChan, prefixedKeys...)
}

// bytesKeyIterator is an iterator returned by ListKeys call.
type bytesKeyIterator struct {
	index int
	len   int
	keys  []string
}

// GetNext returns the following key (+ revision) from the result set.
// When there are no more keys to get, <stop> is returned as *true*
// and <key> and <rev> are default values.
func (ctx *bytesKeyIterator) GetNext() (key string, rev int64, stop bool) {
	if ctx.index >= ctx.len {
		return "", 0, true
	}

	key = string(ctx.keys[ctx.index])
	rev = 0 //ctx.keys[ctx.index].mod
	ctx.index++

	return key, rev, false
}

// Close does nothing since db cursors are not needed.
// The method is required by the code since it implements Iterator API.
func (ctx *bytesKeyIterator) Close() error {
	return nil
}

// bytesKeyValIterator is an iterator returned by ListValues call.
type bytesKeyValIterator struct {
	index int
	len   int
	pairs api.KVPairs
}

// GetNext returns the following item from the result set.
// When there are no more items to get, <stop> is returned as *true* and <val>
// is simply *nil*.
func (ctx *bytesKeyValIterator) GetNext() (val keyval.BytesKeyVal, stop bool) {
	if ctx.index >= ctx.len {
		return nil, true
	}

	key := string(ctx.pairs[ctx.index].Key)
	data := ctx.pairs[ctx.index].Value
	rev := int64(ctx.pairs[ctx.index].ModifyIndex)

	var prevValue []byte
	if len(ctx.pairs) > 0 && ctx.index > 0 {
		prevValue = ctx.pairs[ctx.index-1].Value
	}

	ctx.index++

	return &bytesKeyVal{key, data, prevValue, rev}, false
}

// Close does nothing since db cursors are not needed.
// The method is required by the code since it implements Iterator API.
func (ctx *bytesKeyValIterator) Close() error {
	return nil
}

// bytesKeyVal represents a single key-value pair.
type bytesKeyVal struct {
	key       string
	value     []byte
	prevValue []byte
	revision  int64
}

// Close does nothing since db cursors are not needed.
// The method is required by the code since it implements Iterator API.
func (kv *bytesKeyVal) Close() error {
	return nil
}

// GetValue returns the value of the pair.
func (kv *bytesKeyVal) GetValue() []byte {
	return kv.value
}

// GetPrevValue returns the previous value of the pair.
func (kv *bytesKeyVal) GetPrevValue() []byte {
	return kv.prevValue
}

// GetKey returns the key of the pair.
func (kv *bytesKeyVal) GetKey() string {
	return kv.key
}

// GetRevision returns the revision associated with the pair.
func (kv *bytesKeyVal) GetRevision() int64 {
	return kv.revision
}
