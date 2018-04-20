package consul

import (
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/logging/logrus"
)

type ConsulStore struct {
	client *api.Client
}

func NewConsulStore(addr string) (store *ConsulStore, err error) {
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

	return &ConsulStore{
		client: c,
	}, nil

}

func transformKey(key string) string {
	return strings.TrimPrefix(key, "/")
}

func (c *ConsulStore) Put(key string, data []byte, opts ...datasync.PutOption) error {
	fmt.Printf("put: %q\n", key)
	p := &api.KVPair{Key: transformKey(key), Value: data}
	_, err := c.client.KV().Put(p, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *ConsulStore) NewTxn() keyval.BytesTxn {
	panic("implement me")
}

func (c *ConsulStore) GetValue(key string) (data []byte, found bool, revision int64, err error) {
	fmt.Printf("get value: %q\n", key)
	pair, _, err := c.client.KV().Get(transformKey(key), nil)
	if err != nil {
		return nil, false, 0, err
	} else if pair == nil {
		return nil, false, 0, nil
	}

	return pair.Value, true, int64(pair.ModifyIndex), nil
}

func (c *ConsulStore) ListValues(key string) (keyval.BytesKeyValIterator, error) {
	pairs, _, err := c.client.KV().List(transformKey(key), nil)
	if err != nil {
		return nil, err
	}

	return &bytesKeyValIterator{len: len(pairs), pairs: pairs}, nil
}

func (c *ConsulStore) ListKeys(prefix string) (keyval.BytesKeyIterator, error) {
	keys, _, err := c.client.KV().Keys(transformKey(prefix), "", nil)
	if err != nil {
		return nil, err
	}

	return &bytesKeyIterator{len: len(keys), keys: keys}, nil
}

func (c *ConsulStore) Delete(key string, opts ...datasync.DelOption) (existed bool, err error) {
	fmt.Printf("delete: %q\n", key)
	if _, err := c.client.KV().Delete(transformKey(key), nil); err != nil {
		return false, err
	}

	return true, nil
}

func (c *ConsulStore) Watch(respChan func(keyval.BytesWatchResp), closeChan chan string, keys ...string) error {
	panic("implement me")
}

func (c *ConsulStore) NewBroker(prefix string) keyval.BytesBroker {
	panic("implement me")
}

func (c *ConsulStore) NewWatcher(prefix string) keyval.BytesWatcher {
	panic("implement me")
}

func (c *ConsulStore) Close() error {
	return nil
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
