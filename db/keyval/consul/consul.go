package consul

import (
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
		if err := c.watchKey(resp, closeChan, k); err != nil {
			return err
		}
	}
	return nil
}

func (c *Store) watchKey(resp func(watchResp keyval.BytesWatchResp), closeCh chan string, prefix string) error {
	logrus.DefaultLogger().Debug("WATCH:", prefix)

	// Retrieve KV pairs and latest index
	oldPairs, qm, err := c.client.KV().List(prefix, nil)
	if err != nil {
		return nil
	}

	oldIndex := qm.LastIndex
	oldPairsMap := make(map[string]*api.KVPair)

	logrus.DefaultLogger().Debugf("..waiting: %v old pairs (old index: %v)\n", len(oldPairs), oldIndex)
	for _, pair := range oldPairs {
		logrus.DefaultLogger().Debugf(" - key: %q create: %v modify: %v\n", pair.Key, pair.CreateIndex, pair.ModifyIndex)
		oldPairsMap[pair.Key] = pair
	}

	for {
		select {
		case <-closeCh:
			return nil
		default:
		}

		// Wait for an update to occur since the last index
		var newPairs api.KVPairs
		qOpt := &api.QueryOptions{
			WaitIndex: oldIndex,
			//WaitTime:  time.Second * 2,
		}
		newPairs, qm, err = c.client.KV().List(prefix, qOpt)
		if err != nil {
			return err
		}
		newIndex := qm.LastIndex

		// If the index is same as old one, request probably timed out, so we start again
		if oldIndex == newIndex {
			continue
		}

		logrus.DefaultLogger().Debugf("..waited: %v new pairs (new index: %v) %+v\n", len(newPairs), newIndex, qm)
		for _, pair := range newPairs {
			logrus.DefaultLogger().Debugf(" + key: %q create: %v modify: %v\n", pair.Key, pair.CreateIndex, pair.ModifyIndex)
		}

		var evs []keyval.BytesWatchResp

		// Search for all created and modified KV
		for _, pair := range newPairs {
			if pair.ModifyIndex > oldIndex {
				var prevVal []byte
				if oldPair, ok := oldPairsMap[pair.Key]; ok {
					prevVal = oldPair.Value
				}
				r := etcdv3.NewBytesWatchPutResp(pair.Key, pair.Value, prevVal, int64(pair.ModifyIndex))
				evs = append(evs, r)
			}
			delete(oldPairsMap, pair.Key)
		}
		// Search for all deleted KV
		for _, pair := range oldPairsMap {
			r := etcdv3.NewBytesWatchDelResp(pair.Key, pair.Value, int64(pair.ModifyIndex))
			evs = append(evs, r)
		}

		// Prepare latest KV pairs and last index for next round
		oldIndex = newIndex
		oldPairsMap = make(map[string]*api.KVPair)
		for _, pair := range newPairs {
			oldPairsMap[pair.Key] = pair
		}
		for _, ev := range evs {
			resp(ev)
		}
	}
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
