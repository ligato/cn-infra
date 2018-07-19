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

package bolt

import (
	"fmt"
	"os"
	"strings"
	"github.com/boltdb/bolt"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
)

var boltLogger = logrus.NewLogger("bolt")

func init() {
	if os.Getenv("DEBUG_BOLT_CLIENT") != "" {
		boltLogger.SetLevel(logging.DebugLevel)
	}
}

// Client serves as a client for Bolt KV storage and implements keyval.CoreBrokerWatcher interface.
type Client struct {
	db_path 			*bolt.DB
	bucket_separator 	string
}

func transformKey(key string, separator string) ([][]byte, []byte) {
	if !strings.HasPrefix(key, separator) {
		return nil, nil
	}

	names := strings.Split(key, separator)

	var bucketNames [][]byte 						// := names[:len(names)-1]
	for _, name := range names[1:len(names)-1] {
		bucketNames = append(bucketNames, ([]byte)(name))
	}
	keyInBucket := ([]byte)(names[len(names)-1])

	return bucketNames, keyInBucket
}

// NewTxn creates new transaction
func (client *Client) NewTxn() keyval.BytesTxn {
	tx, _ := client.db_path.Begin(true)
	return &txn{
		readonly: false,
		separator: client.bucket_separator,
		kv: tx,
	}
}

// Create bucket base on given names in current transaction
func createBucket(tx *bolt.Tx, bucketNames [][]byte) (*bolt.Bucket, error) {
	var bucket *bolt.Bucket
	var bucketName []byte
	var errBucket string
	var err error
	for _, bucketName = range bucketNames {
		errBucket = fmt.Sprint("%s//%s", errBucket, bucketName)
		if bucket == nil {
			bucket, err = tx.CreateBucketIfNotExists(bucketName)
		} else {
			bucket, err = bucket.CreateBucketIfNotExists(bucketName)
		}
		if err != nil {
			return nil, /*err*/ fmt.Errorf("Can't create bucket %q!", errBucket)
		}
	}
	return bucket, nil
}

// Find bucket base on given names in current transaction
func findBucket(tx *bolt.Tx, bucketNames [][]byte) (*bolt.Bucket, error) {
	var bucket *bolt.Bucket
	var bucketName []byte
	var errBucket string
	for _, bucketName = range bucketNames {
		errBucket = fmt.Sprint("%s//%s", errBucket, bucketName)
		if bucket == nil {
			bucket = tx.Bucket(bucketName)
		} else {
			bucket = bucket.Bucket(bucketName)
		}
		if bucket == nil {
			return nil, fmt.Errorf("Bucket %q not found!", errBucket)
		}
	}
	return bucket, nil
}

// Put stores given data for the key
func (client *Client) Put(key string, data []byte, opts ...datasync.PutOption) error {
	boltLogger.Debugf("put: (%q,%q)\n", key, data)

	bucketNames, keyInBucket := transformKey(key, client.bucket_separator)
	err := client.db_path.Update(func(tx *bolt.Tx) error {
		bucket, err := createBucket(tx, bucketNames)
		if bucket == nil {
			return err
		} else {
			err = bucket.Put(keyInBucket, data)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// GetValue returns data for the given key
func (client *Client) GetValue(key string) (data []byte, found bool, revision int64, err error) {
	boltLogger.Debugf("get value for key: %q\n", key)

	bucketNames, keyInBucket := transformKey(key, client.bucket_separator)
	var value []byte
	err = client.db_path.View(func(tx *bolt.Tx) error {
		found = false
		bucket, errBucket := findBucket(tx, bucketNames)
		if bucket == nil {
			return fmt.Errorf("Bucket %q not found!", errBucket)
		}
		value = bucket.Get(keyInBucket)
		if value == nil {
			return fmt.Errorf("Value for key %q in bucket %q not found!", keyInBucket, errBucket)
		}
		found = true
		return nil
	})
	return value, found, 0, err
}

// Delete deletes given key
func (client *Client) Delete(key string, opts ...datasync.DelOption) (existed bool, err error) {
	boltLogger.Debugf("delete key: %q\n", key)
	bucketNames, keyInBucket := transformKey(key, client.bucket_separator)
	err = client.db_path.Update(func(tx *bolt.Tx) error {
		bucket, errBucket := findBucket(tx, bucketNames)
		if bucket == nil {
			return fmt.Errorf("Bucket %q not found!", errBucket)
		}
		data := bucket.Get(keyInBucket)
		if  data != nil {
			existed = true
			err = bucket.Delete(keyInBucket)
			return err
		}
		return fmt.Errorf("Key %q not found in bucket %q!", keyInBucket, errBucket)
	})

	return existed, err
}

// Find all keys with given prefix
func (client *Client) findSubtreeKeys(tx *bolt.Tx, bucket *bolt.Bucket, prefix string, keys []string) []string {
	var cursor *bolt.Cursor
	if bucket == nil {
		cursor = tx.Cursor()
	} else {
		cursor = bucket.Cursor()
	}
	for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
		if v == nil {
			parent := prefix
			if prefix == "" {
				prefix = (string)(k)
			} else {
				prefix = fmt.Sprintf( "%s/%s",prefix,(string)(k))
			}
			if bucket == nil {
				keys = client.findSubtreeKeys(tx, tx.Bucket(k), prefix, keys)
			} else {
				keys = client.findSubtreeKeys(tx, bucket.Bucket(k), prefix, keys)
			}
			prefix = parent
		} else {
			keys = append(keys, fmt.Sprintf("%s/%s", prefix, k))
		}
	}
	return keys
}

// Find all keys/value pairs with given prefix
func (client *Client) findSubtreeKeyVals(tx *bolt.Tx, bucket *bolt.Bucket, prefix string, kvps []KVPair) []KVPair {
	var cursor *bolt.Cursor
	if bucket == nil {
		cursor = tx.Cursor()
	} else {
		cursor = bucket.Cursor()
	}
	for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
		if v == nil {
			parent := prefix
			if prefix == "" {
				prefix = (string)(k)
			} else {
				prefix = fmt.Sprintf( "%s/%s",prefix,(string)(k))
			}
			if bucket == nil {
				kvps = client.findSubtreeKeyVals(tx, tx.Bucket(k), prefix, kvps)
			} else {
				kvps = client.findSubtreeKeyVals(tx, bucket.Bucket(k), prefix, kvps)
			}
			prefix = parent
		} else {
			kv := KVPair{fmt.Sprintf("%s/%s", prefix, k), v}
			kvps = append(kvps, kv)
		}
	}
	return kvps
}

// ListKeys returns iterator with keys for given key prefix
func (client *Client) ListKeys(keyPrefix string) (keyval.BytesKeyIterator, error) {
	boltLogger.Debugf("list keys for prefix: %q\n", keyPrefix)

	var prefix string
	var keys []string
	err := client.db_path.View(func(tx *bolt.Tx) error {
		bucketNames, _ := transformKey(keyPrefix, client.bucket_separator)
		bucket, err := findBucket(tx, bucketNames)
		if bucket == nil {
			return fmt.Errorf("Bucket %q not found!", err)
		}
		keys = client.findSubtreeKeys(tx, bucket, prefix, keys)
		return nil
	})
	return &bytesKeyIterator{prefix: keyPrefix, len: len(keys), keys: keys}, err
}

// ListValues returns iterator with key-value pairs for given key prefix
func (client *Client) ListValues(keyPrefix string) (keyval.BytesKeyValIterator, error) {
	boltLogger.Debugf("list values for prefix: %q\n", keyPrefix)

	var prefix string
	var keyVals []KVPair
	err := client.db_path.View(func(tx *bolt.Tx) error {
		bucketNames, _ := transformKey(keyPrefix, client.bucket_separator)
		bucket, err := findBucket(tx, bucketNames)
		if bucket == nil {
			return fmt.Errorf("Bucket %q not found!", err)
		}
		keyVals = client.findSubtreeKeyVals(tx, bucket, prefix, keyVals)
		return err
	})
	return &bytesKeyValIterator{prefix: keyPrefix, len: len(keyVals), pairs: keyVals}, err
}

// Close returns nil.
func (client *Client) Close() error {
	client.db_path.Close()
	return nil
}

// Watch watches given list of key prefixes.
func (client *Client) Watch(resp func(keyval.BytesWatchResp), closeChan chan string, keys ...string) error {
	boltLogger.Debug("Watch:", keys)
	//for _, k := range keys {
	//	if err := c.watch(resp, closeChan, k); err != nil {
	//		return err
	//	}
	//}
	return nil
}
//
//type watchResp struct {
//	typ              datasync.PutDel
//	key              string
//	value, prevValue []byte
//	rev              int64
//}
//
//// GetChangeType returns "Put" for BytesWatchPutResp.
//func (resp *watchResp) GetChangeType() datasync.PutDel {
//	return resp.typ
//}
//
//// GetKey returns the key that the value has been inserted under.
//func (resp *watchResp) GetKey() string {
//	return resp.key
//}
//
//// GetValue returns the value that has been inserted.
//func (resp *watchResp) GetValue() []byte {
//	return resp.value
//}
//
//// GetPrevValue returns the previous value that has been inserted.
//func (resp *watchResp) GetPrevValue() []byte {
//	return resp.prevValue
//}
//
//// GetRevision returns the revision associated with the 'put' operation.
//func (resp *watchResp) GetRevision() int64 {
//	return resp.rev
//}
//
//func (client *Client) watch(resp func(watchResp keyval.BytesWatchResp), closeCh chan string, prefix string) error {
//	boltLogger.Debug("WATCH:", prefix)
//
//	//ctx, cancel := context.WithCancel(context.Background())
//	//
//	//recvChan := c.watchPrefix(ctx, prefix)
//	//
//	//go func(regPrefix string) {
//	//	for {
//	//		select {
//	//		case wr, ok := <-recvChan:
//	//			if !ok {
//	//				boltLogger.WithField("prefix", prefix).Debug("Watch recv chan was closed")
//	//				return
//	//			}
//	//			for _, ev := range wr.Events {
//	//				key := ev.Key
//	//				if !strings.HasPrefix(key, "/") && strings.HasPrefix(regPrefix, "/") {
//	//					key = "/" + key
//	//				}
//	//				var r keyval.BytesWatchResp
//	//				if ev.Type == datasync.Put {
//	//					r = &watchResp{
//	//						typ:       datasync.Put,
//	//						key:       key,
//	//						value:     ev.Value,
//	//						prevValue: ev.PrevValue,
//	//						rev:       ev.Revision,
//	//					}
//	//				} else {
//	//					r = &watchResp{
//	//						typ:   datasync.Delete,
//	//						key:   key,
//	//						value: ev.Value,
//	//						rev:   ev.Revision,
//	//					}
//	//				}
//	//				resp(r)
//	//			}
//	//		case closeVal, ok := <-closeCh:
//	//			if !ok || closeVal == regPrefix {
//	//				boltLogger.WithField("prefix", prefix).Debug("Watch ended")
//	//				cancel()
//	//				return
//	//			}
//	//		}
//	//	}
//	//}(prefix)
//
//	return nil
//}
//
//type watchEvent struct {
//	Type      datasync.PutDel
//	Key       string
//	Value     []byte
//	PrevValue []byte
//	Revision  int64
//}
//
//type watchResponse struct {
//	Events []*watchEvent
//	Err    error
//}
//
//func (client *Client) watchPrefix(ctx context.Context, prefix string) <-chan watchResponse {
//	boltLogger.Debug("watchPrefix:", prefix)
//
//	ch := make(chan watchResponse, 1)
//
//	//// Retrieve KV pairs and latest index
//	//qOpt := &api.QueryOptions{}
//	//oldPairs, qm, err := c.clientC.KV().List(prefix, qOpt.WithContext(ctx))
//	//if err != nil {
//	//	ch <- watchResponse{Err: err}
//	//	close(ch)
//	//	return ch
//	//}
//	//
//	//oldIndex := qm.LastIndex
//	//oldPairsMap := make(map[string]*api.KVPair)
//	//
//	//boltLogger.Debugf("..retrieved: %v old pairs (old index: %v)", len(oldPairs), oldIndex)
//	//for _, pair := range oldPairs {
//	//	boltLogger.Debugf(" - key: %q create: %v modify: %v", pair.Key, pair.CreateIndex, pair.ModifyIndex)
//	//	oldPairsMap[pair.Key] = pair
//	//}
//	//
//	//go func() {
//	//	for {
//	//		// Wait for an update to occur since the last index
//	//		var newPairs api.KVPairs
//	//		qOpt := &api.QueryOptions{
//	//			WaitIndex: oldIndex,
//	//		}
//	//		newPairs, qm, err = c.clientC.KV().List(prefix, qOpt.WithContext(ctx))
//	//		if err != nil {
//	//			ch <- watchResponse{Err: err}
//	//			close(ch)
//	//			return
//	//		}
//	//		newIndex := qm.LastIndex
//	//
//	//		// If the index is same as old one, request probably timed out, so we start again
//	//		if oldIndex == newIndex {
//	//			boltLogger.Debug("index unchanged, next round")
//	//			continue
//	//		}
//	//
//	//		boltLogger.Debugf("prefix %q: %v new pairs (new index: %v) %+v", prefix, len(newPairs), newIndex, qm)
//	//		for _, pair := range newPairs {
//	//			boltLogger.Debugf(" + key: %q create: %v modify: %v", pair.Key, pair.CreateIndex, pair.ModifyIndex)
//	//		}
//	//
//	//		var evs []*watchEvent
//	//
//	//		// Search for all created and modified KV
//	//		for _, pair := range newPairs {
//	//			if pair.ModifyIndex > oldIndex {
//	//				var prevVal []byte
//	//				if oldPair, ok := oldPairsMap[pair.Key]; ok {
//	//					prevVal = oldPair.Value
//	//				}
//	//				evs = append(evs, &watchEvent{
//	//					Type:      datasync.Put,
//	//					Key:       pair.Key,
//	//					Value:     pair.Value,
//	//					PrevValue: prevVal,
//	//					Revision:  int64(pair.ModifyIndex),
//	//				})
//	//			}
//	//			delete(oldPairsMap, pair.Key)
//	//		}
//	//		// Search for all deleted KV
//	//		for _, pair := range oldPairsMap {
//	//			evs = append(evs, &watchEvent{
//	//				Type:      datasync.Delete,
//	//				Key:       pair.Key,
//	//				PrevValue: pair.Value,
//	//				Revision:  int64(pair.ModifyIndex),
//	//			})
//	//		}
//	//
//	//		// Prepare latest KV pairs and last index for next round
//	//		oldIndex = newIndex
//	//		oldPairsMap = make(map[string]*api.KVPair)
//	//		for _, pair := range newPairs {
//	//			oldPairsMap[pair.Key] = pair
//	//		}
//	//
//	//		ch <- watchResponse{Events: evs}
//	//	}
//	//}()
//	return ch
//}

// NewBroker creates a new instance of a proxy that provides
// access to etcd. The proxy will reuse the connection from Client.
// <prefix> will be prepended to the key argument in all calls from the created
// BrokerWatcher. To avoid using a prefix, pass keyval. Root constant as
// an argument.
func (client *Client) NewBroker(prefix string) keyval.BytesBroker {
	return &BrokerWatcher{
		Client: client,
		prefix: prefix,
	}
}

// NewWatcher creates a new instance of a proxy that provides
// access to etcd. The proxy will reuse the connection from Client.
// <prefix> will be prepended to the key argument in all calls on created
// BrokerWatcher. To avoid using a prefix, pass keyval. Root constant as
// an argument.
func (client *Client) NewWatcher(prefix string) keyval.BytesWatcher {
	return &BrokerWatcher{
		Client: client,
		prefix: prefix,
	}
}

// BrokerWatcher uses Client to access the datastore.
// The connection can be shared among multiple BrokerWatcher.
// In case of accessing a particular subtree in Consul only,
// BrokerWatcher allows defining a keyPrefix that is prepended
// to all keys in its methods in order to shorten keys used in arguments.
type BrokerWatcher struct {
	*Client
	prefix string
}

//func (pdb *BrokerWatcher) prefixKey(key string) string {
//	return filepath.Join(pdb.prefix, key)
//}
//
//// Put calls 'Put' function of the underlying BytesConnectionEtcd.
//// KeyPrefix defined in constructor is prepended to the key argument.
//func (pdb *BrokerWatcher) Put(key string, data []byte, opts ...datasync.PutOption) error {
//	return pdb.Client.Put(pdb.prefixKey(key), data, opts...)
//}
//
//// NewTxn creates a new transaction.
//// KeyPrefix defined in constructor will be prepended to all key arguments
//// in the transaction.
//func (pdb *BrokerWatcher) NewTxn() keyval.BytesTxn {
//	return pdb.Client.NewTxn()
//}
//
//// GetValue calls 'GetValue' function of the underlying BytesConnectionEtcd.
//// KeyPrefix defined in constructor is prepended to the key argument.
//func (pdb *BrokerWatcher) GetValue(key string) (data []byte, found bool, revision int64, err error) {
//	return pdb.Client.GetValue(pdb.prefixKey(key))
//}
//
//// ListValues calls 'ListValues' function of the underlying BytesConnectionEtcd.
//// KeyPrefix defined in constructor is prepended to the key argument.
//// The prefix is removed from the keys of the returned values.
//func (pdb *BrokerWatcher) ListValues(key string) (keyval.BytesKeyValIterator, error) {
//	//pairs, _, err := pdb.clientC.KV().List(pdb.prefixKey(transformKey(key)), nil)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	//return &bytesKeyValIterator{len: len(pairs), pairs: pairs, prefix: pdb.prefix}, nil
//	return nil, nil
//}
//
//// ListKeys calls 'ListKeys' function of the underlying BytesConnectionEtcd.
//// KeyPrefix defined in constructor is prepended to the argument.
//func (pdb *BrokerWatcher) ListKeys(prefix string) (keyval.BytesKeyIterator, error) {
//	//keys, qm, err := pdb.clientC.KV().Keys(pdb.prefixKey(transformKey(prefix)), "", nil)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	//return &bytesKeyIterator{len: len(keys), keys: keys, prefix: pdb.prefix, lastIndex: qm.LastIndex}, nil
//	return nil, nil
//}
//
//// Delete calls 'Delete' function of the underlying BytesConnectionEtcd.
//// KeyPrefix defined in constructor is prepended to the key argument.
//func (pdb *BrokerWatcher) Delete(key string, opts ...datasync.DelOption) (existed bool, err error) {
//	return pdb.Client.Delete(pdb.prefixKey(key), opts...)
//}
//
//// Watch starts subscription for changes associated with the selected <keys>.
//// KeyPrefix defined in constructor is prepended to all <keys> in the argument
//// list. The prefix is removed from the keys returned in watch events.
//// Watch events will be delivered to <resp> callback.
//func (pdb *BrokerWatcher) Watch(resp func(keyval.BytesWatchResp), closeChan chan string, keys ...string) error {
//	var prefixedKeys []string
//	for _, key := range keys {
//		prefixedKeys = append(prefixedKeys, pdb.prefixKey(key))
//	}
//	return pdb.Client.Watch(func(origResp keyval.BytesWatchResp) {
//		r := origResp.(*watchResp)
//		r.key = strings.TrimPrefix(r.key, pdb.prefix)
//		resp(r)
//	}, closeChan, prefixedKeys...)
//}

// KVPair is used to represent a single K/V entry
type KVPair struct {
	// Key is the name of the key.
	Key string

	// Value is the value for the key.
	Value []byte
}

// bytesKeyIterator is an iterator returned by ListKeys call.
type bytesKeyIterator struct {
	prefix    string
	index     int
	len       int
	keys      []string
}

// bytesKeyValIterator is an iterator returned by ListValues call.
type bytesKeyValIterator struct {
	prefix    string
	index     int
	len       int
	pairs     []KVPair
}

// bytesKeyVal represents a single key-value pair.
type bytesKeyVal struct {
	key       string
	value     []byte
	prevValue []byte
	revision  int64
}

// GetNext returns the following key (+ revision) from the result set.
// When there are no more keys to get, <stop> is returned as *true*
// and <key> and <rev> are default values.
func (it *bytesKeyIterator) GetNext() (key string, rev int64, stop bool) {
	if it.index >= it.len {
		return "", 0, true
	}

	key = string(it.keys[it.index])
	if it.prefix != "" {
		key = strings.TrimPrefix(key, it.prefix)
	}
	it.index++

	return key, 0, false
}

// Close does nothing since db cursors are not needed.
// The method is required by the code since it implements Iterator API.
func (it *bytesKeyIterator) Close() error {
	return nil
}

// GetNext returns the following item from the result set.
// When there are no more items to get, <stop> is returned as *true* and <val>
// is simply *nil*.
func (it *bytesKeyValIterator) GetNext() (val keyval.BytesKeyVal, stop bool) {
	if it.index >= it.len {
		return nil, true
	}

	key := string(it.pairs[it.index].Key)
	if it.prefix != "" {
		key = strings.TrimPrefix(key, it.prefix)
	}
	data := it.pairs[it.index].Value

	var prevValue []byte
	if len(it.pairs) > 0 && it.index > 0 {
		prevValue = it.pairs[it.index-1].Value
	}

	it.index++

	return &bytesKeyVal{key, data, prevValue, 0}, false
}

// Close does nothing since db cursors are not needed.
// The method is required by the code since it implements Iterator API.
func (it *bytesKeyValIterator) Close() error {
	return nil
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
