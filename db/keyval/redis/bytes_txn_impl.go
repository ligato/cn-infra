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

package redis

import (
	"fmt"
	"strings"

	goredis "github.com/go-redis/redis"
	"github.com/howeyc/crc16"
	"github.com/ligato/cn-infra/db/keyval"
)

type op struct {
	key   string
	value []byte
	del   bool
}

type Txn struct {
	db     *BytesConnectionRedis
	ops    []op
	prefix string
}

func (tx *Txn) addPrefix(key string) string {
	return tx.prefix + key
}

// Put adds a new 'put' operation to a previously created transaction.
// If the key does not exist in the data store, a new key-value item
// will be added to the data store. If key exists in the data store,
// the existing value will be overwritten with the value from this
// operation.
func (tx *Txn) Put(key string, value []byte) keyval.BytesTxn {
	tx.ops = append(tx.ops, op{tx.addPrefix(key), value, false})
	return tx
}

// Delete adds a new 'delete' operation to a previously created
// transaction.
func (tx *Txn) Delete(key string) keyval.BytesTxn {
	tx.ops = append(tx.ops, op{tx.addPrefix(key), nil, true})
	return tx
}

// Commit commits all operations in a transaction to the data store.
// Commit is atomic - either all operations in the transaction are
// committed to the data store, or none of them.
func (tx *Txn) Commit() (err error) {
	if tx.db.closed {
		return fmt.Errorf("Commit() called on a closed connection")
	}
	tx.db.Debug("Commit()")

	if len(tx.ops) == 0 {
		return nil
	}

	// redigo
	if tx.db.pool != nil {
		toBeDeleted := []interface{}{}
		msetArgs := []interface{}{}
		for _, op := range tx.ops {
			if op.del {
				toBeDeleted = append(toBeDeleted, op.key)
			} else {
				msetArgs = append(msetArgs, op.key)
				msetArgs = append(msetArgs, string(op.value))
			}
		}

		conn := tx.db.pool.Get()
		defer conn.Close()

		if len(toBeDeleted) > 0 {
			_, err = conn.Do("DEL", toBeDeleted...)
			if err != nil {
				return fmt.Errorf("Do(DEL) failed: %s", err)
			}
		}
		if len(msetArgs) > 0 {
			_, err = conn.Do("MSET", msetArgs...)
			if err != nil {
				return fmt.Errorf("Do(MSET) failed: %s", err)
			}
		}
		return nil
	}

	// go-redis

	pipeline := tx.db.client.TxPipeline()
	for _, op := range tx.ops {
		if op.del {
			pipeline.Del(op.key)
		} else {
			pipeline.Set(op.key, op.value, 0)
		}
	}
	_, err = pipeline.Exec()
	if err != nil {
		// Redis cluster won't let you run multi-key commands in case of cross slot.
		// - Cross slot check may be useful indicator in case of failure.
		if _, yes := tx.db.client.(*goredis.ClusterClient); yes {
			checkCrossSlot(tx)
		}
		return fmt.Errorf("%T.Exec() failed: %s", pipeline, err)
	}
	return nil
}

// CROSSSLOT Keys in request don't hash to the same slot
// https://stackoverflow.com/questions/38042629/redis-cross-slot-error
// https://redis.io/topics/cluster-spec#keys-hash-tags
// https://redis.io/topics/cluster-tutorial
// "Redis Cluster supports multiple key operations as long as all the keys involved into a single
// command execution (or whole transaction, or Lua script execution) all belong to the same hash
// slot. The user can force multiple keys to be part of the same hash slot by using a concept
// called hash tags."
func checkCrossSlot(tx *Txn) bool {
	var hashSlot uint16
	var key string

	for _, op := range tx.ops {
		if hashSlot == 0 {
			hashSlot = getHashSlot(op.key)
			key = op.key
		} else {
			slot := getHashSlot(op.key)
			if slot != hashSlot {
				tx.db.Warnf("%T: Found CROSS SLOT keys (%s, %d) and (%s, %d)",
					*tx, key, hashSlot, op.key, slot)
				return true
			}
		}
	}
	return false
}

func getHashSlot(key string) uint16 {
	start := strings.Index(key, "{")
	if start != -1 {
		start++
		end := strings.Index(key[start:], "}")
		if end != -1 {
			key = key[start:end]
		}
	}
	const redisHashSlotCount = 16384
	return crc16.ChecksumCCITT([]byte(key)) % redisHashSlotCount
}
