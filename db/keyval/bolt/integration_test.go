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

package bolt

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	. "github.com/onsi/gomega"

	"go.ligato.io/cn-infra/v2/datasync"
	"go.ligato.io/cn-infra/v2/db/keyval"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/logging/logs"
)

func init() {
	logs.DefaultLogger().SetLevel(logging.DebugLevel)
	boltLogger.SetLevel(logging.DebugLevel)
}

type testCtx struct {
	*testing.T
	client *Client
}

const testDbPath = "/tmp/bolt.db"

func setupTest(t *testing.T, newDB bool) *testCtx {
	RegisterTestingT(t)

	if newDB {
		err := os.Remove(testDbPath)
		if err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
			return nil
		}
	}

	client, err := NewClient(&Config{
		DbPath:   testDbPath,
		FileMode: 432,
	})
	Expect(err).ToNot(HaveOccurred())
	if err != nil {
		return nil
	}

	return &testCtx{T: t, client: client}
}

func (tc *testCtx) teardownTest() {
	Expect(tc.client.Close()).To(Succeed())
}

func (tc *testCtx) isInDB(key string, expectedVal []byte) (exists bool) {
	err := tc.client.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(rootBucket)
		if val := b.Get([]byte(key)); val != nil {
			exists = true
		}
		return nil
	})
	if err != nil {
		tc.Fatal(err)
	}
	return
}

func (tc *testCtx) populateDB(data map[string][]byte) {
	err := tc.client.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(rootBucket)
		for key, val := range data {
			if err := b.Put([]byte(key), val); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		tc.Fatal(err)
	}
	return
}

func TestPut(t *testing.T) {
	tc := setupTest(t, true)
	defer tc.teardownTest()

	var key = "/agent/agent1/config/interface/iface0"
	var val = []byte("val")

	err := tc.client.Put(key, val)
	Expect(err).ToNot(HaveOccurred())
	Expect(tc.isInDB(key, val)).To(BeTrue())
}

func TestGet(t *testing.T) {
	tc := setupTest(t, true)
	defer tc.teardownTest()

	var key = "/agent/agent1/config/interface/iface0"
	var val = []byte("val")

	err := tc.client.Put(key, val)
	Expect(err).ToNot(HaveOccurred())
	Expect(tc.isInDB(key, val)).To(BeTrue())
}

func TestListKeys(t *testing.T) {
	tc := setupTest(t, true)
	defer tc.teardownTest()

	tc.populateDB(map[string][]byte{
		"/my/key/1":    []byte("val1"),
		"/my/key/2":    []byte("val2"),
		"/other/key/0": []byte("val0"),
	})

	kvi, err := tc.client.ListKeys("/my/key/")
	Expect(err).ToNot(HaveOccurred())
	Expect(kvi).NotTo(BeNil())

	expectedKeys := []string{"/my/key/1", "/my/key/2"}
	for i := 0; i <= len(expectedKeys); i++ {
		key, _, all := kvi.GetNext()
		if i == len(expectedKeys) {
			Expect(all).To(BeTrue())
			break
		}
		Expect(all).To(BeFalse())
		Expect(key).To(BeEquivalentTo(expectedKeys[i]))
	}
}

func TestListValues(t *testing.T) {
	tc := setupTest(t, true)
	defer tc.teardownTest()

	tc.populateDB(map[string][]byte{
		"/my/key/1":    []byte("val1"),
		"/my/key/2":    []byte("val2"),
		"/other/key/0": []byte("val0"),
	})

	kvi, err := tc.client.ListValues("/my/key/")
	Expect(err).ToNot(HaveOccurred())
	Expect(kvi).NotTo(BeNil())

	expectedKeys := []string{"/my/key/1", "/my/key/2"}
	expectedValues := [][]byte{[]byte("val1"), []byte("val2")}
	for i := 0; i <= len(expectedKeys); i++ {
		kv, all := kvi.GetNext()
		if i == len(expectedKeys) {
			Expect(all).To(BeTrue())
			break
		}
		Expect(all).To(BeFalse())
		Expect(kv.GetKey()).To(BeEquivalentTo(expectedKeys[i]))
		Expect(bytes.Compare(kv.GetValue(), expectedValues[i])).To(BeZero())
	}
}

func TestListKeysBroker(t *testing.T) {
	tc := setupTest(t, true)
	defer tc.teardownTest()

	tc.populateDB(map[string][]byte{
		"/my/key/1":    []byte("val1"),
		"/my/key/2":    []byte("val2"),
		"/my/keyx/xx":  []byte("x"),
		"/my/xkey/xx":  []byte("x"),
		"/other/key/0": []byte("val0"),
	})

	broker := tc.client.NewBroker("/my/")
	kvi, err := broker.ListKeys("key/")
	Expect(err).ToNot(HaveOccurred())
	Expect(kvi).NotTo(BeNil())

	expectedKeys := []string{"key/1", "key/2"}
	for i := 0; i <= len(expectedKeys); i++ {
		key, _, all := kvi.GetNext()
		if i == len(expectedKeys) {
			Expect(all).To(BeTrue())
			break
		}
		Expect(all).To(BeFalse())
		Expect(key).To(BeEquivalentTo(expectedKeys[i]))
	}
}

func TestListValuesBroker(t *testing.T) {
	tc := setupTest(t, true)
	defer tc.teardownTest()

	tc.populateDB(map[string][]byte{
		"/my/key/1":    []byte("val1"),
		"/my/key/2":    []byte("val2"),
		"/my/keyx/xx":  []byte("x"),
		"/my/xkey/xx":  []byte("x"),
		"/other/key/0": []byte("val0"),
	})

	broker := tc.client.NewBroker("/my/")
	kvi, err := broker.ListValues("key/")
	Expect(err).ToNot(HaveOccurred())
	Expect(kvi).NotTo(BeNil())

	expectedKeys := []string{"key/1", "key/2"}
	for i := 0; i <= len(expectedKeys); i++ {
		kv, all := kvi.GetNext()
		if i == len(expectedKeys) {
			Expect(all).To(BeTrue())
			break
		}
		Expect(all).To(BeFalse())
		Expect(kv.GetKey()).To(BeEquivalentTo(expectedKeys[i]))
	}
}

func TestDelete(t *testing.T) {
	tc := setupTest(t, true)
	defer tc.teardownTest()

	var key = "/agent/agent1/config/interface/iface0"
	var val = []byte("val")

	err := tc.client.Put(key, val)
	Expect(err).ToNot(HaveOccurred())
	existed, err := tc.client.Delete(key)
	Expect(err).ToNot(HaveOccurred())
	Expect(existed).To(BeTrue())

	existed, err = tc.client.Delete("/this/key/does/not/exists")
	Expect(err).To(HaveOccurred())
	Expect(existed).To(BeFalse())
	Expect(tc.isInDB(key, val)).To(BeFalse())
}

func TestPutInTxn(t *testing.T) {
	tc := setupTest(t, true)
	defer tc.teardownTest()

	txn := tc.client.NewTxn()
	Expect(txn).ToNot(BeNil())

	var key1 = "/agent/agent1/config/interface/iface0"
	var val1 = []byte("iface0")
	var key2 = "/agent/agent1/config/interface/iface1"
	var val2 = []byte("iface1")
	var key3 = "/agent/agent1/config/interface/iface2"
	var val3 = []byte("iface2")

	txn.Put(key1, val1).
		Put(key2, val2).
		Put(key3, val3)
	Expect(txn.Commit(context.Background())).To(Succeed())
	Expect(tc.isInDB(key1, val1)).To(BeTrue())
	Expect(tc.isInDB(key2, val2)).To(BeTrue())
	Expect(tc.isInDB(key3, val3)).To(BeTrue())
}

func TestDeleteInTxn(t *testing.T) {
	tc := setupTest(t, true)
	defer tc.teardownTest()

	txn := tc.client.NewTxn()
	Expect(txn).ToNot(BeNil())

	var key1 = "/agent/agent1/config/interface/iface0"
	var val1 = []byte("iface0")
	var key2 = "/agent/agent1/config/interface/iface1"
	var val2 = []byte("iface1")
	var key3 = "/agent/agent1/config/interface/iface2"
	var val3 = []byte("iface2")

	txn.Put(key1, val1).
		Put(key2, val2).
		Put(key3, val3).
		Delete(key2)
	Expect(txn.Commit(context.Background())).To(Succeed())
	Expect(tc.isInDB(key1, val1)).To(BeTrue())
	Expect(tc.isInDB(key2, val2)).To(BeFalse())
	Expect(tc.isInDB(key3, val3)).To(BeTrue())
}

func TestWatchPut(t *testing.T) {
	ctx := setupTest(t, true)
	defer ctx.teardownTest()

	const watchPrefix = "/key/"
	const watchKey = watchPrefix + "val1"

	closeCh := make(chan string)
	watchCh := make(chan keyval.BytesWatchResp)
	err := ctx.client.Watch(keyval.ToChan(watchCh), closeCh, watchPrefix)
	Expect(err).To(BeNil())

	Expect(ctx.client.Put("/something/else/val1", []byte{0, 0, 7})).To(Succeed())
	Expect(ctx.client.Put(watchKey, []byte{1, 2, 3})).To(Succeed())

	var resp keyval.BytesWatchResp
	Eventually(watchCh).Should(Receive(&resp))
	Expect(resp.GetKey()).Should(Equal(watchKey))
	Expect(resp.GetValue()).Should(Equal([]byte{1, 2, 3}))
	Expect(resp.GetPrevValue()).Should(BeNil())
	Expect(resp.GetChangeType()).Should(Equal(datasync.Put))
}

func TestWatchDel(t *testing.T) {
	ctx := setupTest(t, true)
	defer ctx.teardownTest()

	const watchPrefix = "/key/"
	const watchKey = watchPrefix + "val1"

	Expect(ctx.client.Put("/something/else/val1", []byte{0, 0, 7})).To(Succeed())
	Expect(ctx.client.Put(watchKey, []byte{1, 2, 3})).To(Succeed())

	closeCh := make(chan string)
	watchCh := make(chan keyval.BytesWatchResp)
	err := ctx.client.Watch(keyval.ToChan(watchCh), closeCh, watchKey)
	Expect(err).To(BeNil())

	Expect(ctx.client.Delete("/something/else/val1")).To(BeTrue())
	Expect(ctx.client.Delete(watchKey)).To(BeTrue())

	var resp keyval.BytesWatchResp
	Eventually(watchCh).Should(Receive(&resp))
	Expect(resp.GetKey()).Should(Equal(watchKey))
	Expect(resp.GetValue()).Should(BeNil())
	Expect(resp.GetPrevValue()).Should(Equal([]byte{1, 2, 3}))
	Expect(resp.GetChangeType()).Should(Equal(datasync.Delete))
}

func TestWatchPutBroker(t *testing.T) {
	ctx := setupTest(t, true)
	defer ctx.teardownTest()

	const brokerPrefix = "/my/prefix/"
	const watchPrefix = "key/"
	const watchKey = watchPrefix + "val1"

	broker := ctx.client.NewWatcher(brokerPrefix)

	closeCh := make(chan string)
	watchCh := make(chan keyval.BytesWatchResp)

	err := broker.Watch(keyval.ToChan(watchCh), closeCh, watchPrefix)
	Expect(err).To(BeNil())

	Expect(ctx.client.Put(brokerPrefix+"something/else/val1", []byte{0, 0, 7})).To(Succeed())
	Expect(ctx.client.Put(brokerPrefix+watchKey, []byte{1, 2, 3})).To(Succeed())

	var resp keyval.BytesWatchResp
	Eventually(watchCh).Should(Receive(&resp))
	Expect(resp.GetKey()).Should(Equal(watchKey))
	Expect(resp.GetValue()).Should(Equal([]byte{1, 2, 3}))
	Expect(resp.GetPrevValue()).Should(BeNil())
	Expect(resp.GetChangeType()).Should(Equal(datasync.Put))
}

func TestFilterDupNotifs(t *testing.T) {
	ctx := setupTest(t, true)
	ctx.client.cfg.FilterDupNotifs = true
	defer ctx.teardownTest()

	const brokerPrefix = "/my/prefix/"
	const watchPrefix = "key/"
	const watchKey = watchPrefix + "val1"

	broker := ctx.client.NewWatcher(brokerPrefix)

	closeCh := make(chan string)
	watchCh := make(chan keyval.BytesWatchResp, 5)

	err := broker.Watch(keyval.ToChan(watchCh), closeCh, watchPrefix)
	Expect(err).To(BeNil())

	for i := 0; i < 10; i++ {
		Expect(ctx.client.Put(brokerPrefix+watchKey, []byte{1, 2, 3})).To(Succeed())
	}

	var resp keyval.BytesWatchResp
	Eventually(watchCh).Should(Receive(&resp))
	Expect(resp.GetKey()).Should(Equal(watchKey))
	Expect(resp.GetValue()).Should(Equal([]byte{1, 2, 3}))
	Expect(resp.GetPrevValue()).Should(BeNil())
	Expect(resp.GetChangeType()).Should(Equal(datasync.Put))
	Consistently(watchCh, time.Second).ShouldNot(Receive()) // filter duplicate notifications

	// put different data
	Expect(ctx.client.Put(brokerPrefix+watchKey, []byte{1, 2, 3, 4})).To(Succeed())
	Eventually(watchCh).Should(Receive(&resp))
	Expect(resp.GetKey()).Should(Equal(watchKey))
	Expect(resp.GetValue()).Should(Equal([]byte{1, 2, 3, 4}))
	Expect(resp.GetPrevValue()).Should(Equal([]byte{1, 2, 3}))
	Expect(resp.GetChangeType()).Should(Equal(datasync.Put))

	// delete "twice"
	Expect(ctx.client.Delete(brokerPrefix + watchKey)).To(BeTrue())
	existed, _ := ctx.client.Delete(brokerPrefix + watchKey)
	Expect(existed).To(BeFalse())
	Eventually(watchCh).Should(Receive(&resp))
	Expect(resp.GetKey()).Should(Equal(watchKey))
	Expect(resp.GetValue()).Should(BeNil())
	Expect(resp.GetPrevValue()).Should(Equal([]byte{1, 2, 3, 4}))
	Expect(resp.GetChangeType()).Should(Equal(datasync.Delete))
	Consistently(watchCh, time.Second).ShouldNot(Receive()) // filter duplicate notifications
}

func TestClosedWatch(t *testing.T) {
	ctx := setupTest(t, true)
	defer ctx.teardownTest()

	const watchPrefix = "/prefix/"
	const watchKey1 = "key1"
	const watchKey2 = "key2"

	closeCh := make(chan string)
	watchCh := make(chan keyval.BytesWatchResp)
	broker := ctx.client.NewBroker(watchPrefix)
	watcher := ctx.client.NewWatcher(watchPrefix)
	err := watcher.Watch(keyval.ToChan(watchCh), closeCh, watchKey1, watchKey2)
	Expect(err).To(BeNil())

	var resp keyval.BytesWatchResp
	Expect(broker.Put(watchKey1, []byte{1, 2, 3})).To(Succeed())
	Eventually(watchCh).Should(Receive(&resp))
	Expect(resp.GetKey()).Should(Equal(watchKey1))
	Expect(resp.GetValue()).Should(Equal([]byte{1, 2, 3}))
	Expect(resp.GetPrevValue()).Should(BeNil())
	Expect(resp.GetChangeType()).Should(Equal(datasync.Put))

	// close watch for watchKey1 but not for watchKey2 yet
	closeCh <- watchKey1
	time.Sleep(time.Second)
	Expect(broker.Put(watchKey1, []byte{4, 5, 6})).To(Succeed())
	Consistently(watchCh).ShouldNot(Receive())

	Expect(broker.Put(watchKey2, []byte{1, 2, 3})).To(Succeed())
	Eventually(watchCh).Should(Receive(&resp))
	Expect(resp.GetKey()).Should(Equal(watchKey2))
	Expect(resp.GetValue()).Should(Equal([]byte{1, 2, 3}))
	Expect(resp.GetPrevValue()).Should(BeNil())
	Expect(resp.GetChangeType()).Should(Equal(datasync.Put))

	// close watching completely
	close(closeCh)
	time.Sleep(time.Second)
	Expect(broker.Put(watchKey1, []byte{4, 5, 6})).To(Succeed())
	Consistently(watchCh).ShouldNot(Receive())
	Expect(broker.Put(watchKey2, []byte{4, 5, 6})).To(Succeed())
	Consistently(watchCh).ShouldNot(Receive())
}

func TestReuseCloseChannel(t *testing.T) {
	ctx := setupTest(t, true)
	defer ctx.teardownTest()

	const watchPrefix = "/prefix/"
	const watchKey1 = "key1"
	const watchKey2 = "key2"

	closeCh := make(chan string)
	watchCh := make(chan keyval.BytesWatchResp)
	broker := ctx.client.NewBroker(watchPrefix)
	watcher := ctx.client.NewWatcher(watchPrefix)
	err := watcher.Watch(keyval.ToChan(watchCh), closeCh, watchKey1)
	Expect(err).To(BeNil())

	var resp keyval.BytesWatchResp
	Expect(broker.Put(watchKey1, []byte{1, 2, 3})).To(Succeed())
	Eventually(watchCh).Should(Receive(&resp))
	Expect(resp.GetKey()).Should(Equal(watchKey1))
	Expect(resp.GetValue()).Should(Equal([]byte{1, 2, 3}))
	Expect(resp.GetPrevValue()).Should(BeNil())
	Expect(resp.GetChangeType()).Should(Equal(datasync.Put))

	// watchKey2 is not watched yet
	Expect(broker.Put(watchKey2, []byte{1, 2, 3})).To(Succeed())
	Consistently(watchCh).ShouldNot(Receive())

	// add watchPrefix to the list of watched key prefixes
	err = watcher.Watch(keyval.ToChan(watchCh), closeCh, "")
	Expect(err).To(BeNil())
	time.Sleep(time.Second)

	// change both keys - watch callbacks are called one after another from one
	// common go routine
	Expect(broker.Put(watchKey2, []byte{4, 5, 6})).To(Succeed())
	Expect(broker.Put(watchKey1, []byte{4, 5, 6})).To(Succeed())

	Eventually(watchCh).Should(Receive(&resp))
	Expect(resp.GetKey()).Should(Equal(watchKey2))
	Expect(resp.GetValue()).Should(Equal([]byte{4, 5, 6}))
	Expect(resp.GetPrevValue()).Should(Equal([]byte{1, 2, 3}))
	Expect(resp.GetChangeType()).Should(Equal(datasync.Put))

	Eventually(watchCh).Should(Receive(&resp))
	Expect(resp.GetKey()).Should(Equal(watchKey1))
	Expect(resp.GetValue()).Should(Equal([]byte{4, 5, 6}))
	Expect(resp.GetPrevValue()).Should(Equal([]byte{1, 2, 3}))
	Expect(resp.GetChangeType()).Should(Equal(datasync.Put))

	// close watch for previously added watchPrefix
	closeCh <- "" // watcher is prefixed, therefore empty string translates to watchPrefix
	Expect(broker.Put(watchKey2, []byte{7, 8, 9})).To(Succeed())
	Expect(broker.Put(watchKey1, []byte{7, 8, 9})).To(Succeed())

	Eventually(watchCh).Should(Receive(&resp))
	Expect(resp.GetKey()).Should(Equal(watchKey1))
	Expect(resp.GetValue()).Should(Equal([]byte{7, 8, 9}))
	Expect(resp.GetPrevValue()).Should(Equal([]byte{4, 5, 6}))
	Expect(resp.GetChangeType()).Should(Equal(datasync.Put))

	Consistently(watchCh).ShouldNot(Receive())
}
