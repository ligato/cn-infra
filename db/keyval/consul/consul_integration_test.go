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
	"sync"
	"testing"
	"time"

	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	. "github.com/onsi/gomega"
)

func init() {
	logrus.DefaultLogger().SetLevel(logging.DebugLevel)
}

type testCtx struct {
	store *Store
}

func setupTest(t *testing.T) *testCtx {
	RegisterTestingT(t)

	store, err := NewConsulStore("127.0.0.1:8500")
	if err != nil {
		t.Fatal("connecting to consul failed:", err)
	}

	return &testCtx{store}
}

func (ctx *testCtx) teardownTest() {
	ctx.store.Close()
}

func TestPut(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	err := ctx.store.Put("key", []byte("val"))
	Expect(err).ToNot(HaveOccurred())
}

func TestGetValue(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	data, found, rev, err := ctx.store.GetValue("key")
	Expect(err).ToNot(HaveOccurred())
	Expect(data).To(Equal([]byte("val")))
	Expect(found).To(BeTrue())
	Expect(rev).NotTo(BeZero())
}

func TestDelete(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	existed, err := ctx.store.Delete("key")
	Expect(err).ToNot(HaveOccurred())
	Expect(existed).To(BeTrue())
}

func TestWatch(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	watchKey := "key/"

	closeCh := make(chan string)
	watchCh := make(chan keyval.BytesWatchResp)
	err := ctx.store.Watch(keyval.ToChan(watchCh), closeCh, watchKey)
	Expect(err).To(BeNil())

	var wg sync.WaitGroup
	wg.Add(1)

	go func(expectedKey string) {
		select {
		case resp := <-watchCh:
			Expect(resp).NotTo(BeNil())
			Expect(resp.GetKey()).To(BeEquivalentTo(expectedKey))
		case <-time.After(time.Second):
			t.Error("Watch resp not received")
			t.FailNow()
		}
		close(closeCh)
		wg.Done()
	}(watchKey + "val1")

	ctx.store.Put("/something/else/val1", []byte{0, 0, 7})
	ctx.store.Put(watchKey+"val1", []byte{1, 2, 3})

	wg.Wait()

	time.Sleep(time.Second)
}
