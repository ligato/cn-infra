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

package bolt

import (
	"bytes"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	. "github.com/onsi/gomega"
)

func init() {
	logrus.DefaultLogger().SetLevel(logging.DebugLevel)
}

const DbPath = "/tmp/bolt.db"
const BucketSeparator = "/"

func setupTest(t *testing.T, newDB bool) *Client {
	RegisterTestingT(t)

	if newDB {
		if err := os.Remove(DbPath); err != nil {
			if !os.IsNotExist(err) {
				return nil
			}
		}
	}

	var err error
	client := &Client{}
	client.dbPath, err = bolt.Open(DbPath, 432, nil)
	if err != nil {
		return nil
	}
	client.bucketSeparator = BucketSeparator

	return client
}

func (client *Client) teardownTest() {
	client.Close()
}

func (client *Client) checkIfExists(key string, expectedVal []byte) bool {
	_, found, _, _ := client.GetValue(key)
	return found
}

func TestPut(t *testing.T) {
	client := setupTest(t, true)
	defer client.teardownTest()

	var key = "/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop0"
	err := client.Put(key, []byte("val"))
	Expect(err).ToNot(HaveOccurred())
	Expect(client.checkIfExists(key, []byte("val"))).To(BeTrue())
}

func TestGet(t *testing.T) {
	client := setupTest(t, true)
	defer client.teardownTest()

	var key = "/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop1"
	var value = ([]byte)("val")

	err := client.Put(key, value)
	Expect(err).ToNot(HaveOccurred())
	data, found, _, err := client.GetValue(key)
	Expect(found).To(BeTrue())
	Expect(bytes.Equal(data, value)).To(BeTrue())
}

func TestDelete(t *testing.T) {
	client := setupTest(t, true)
	defer client.teardownTest()

	var key = "/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop2"
	var value = ([]byte)("val")

	err := client.Put(key, value)
	Expect(err).ToNot(HaveOccurred())
	existed, err := client.Delete(key)
	Expect(err).ToNot(HaveOccurred())
	Expect(existed).To(BeTrue())

	existed, err = client.Delete("/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop")
	Expect(err).To(HaveOccurred())
	Expect(existed).To(BeFalse())
	Expect(client.checkIfExists(key, []byte("val"))).To(BeFalse())
}

func TestPutInTxn(t *testing.T) {
	client := setupTest(t, true)
	defer client.teardownTest()

	txn := client.NewTxn()
	Expect(txn).ToNot(BeNil())

	var key1 = "/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop0"
	var value1 = ([]byte)("bvi_loop1")
	var key2 = "/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop1"
	var value2 = ([]byte)("bvi_loop2")
	var key3 = "/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop2"
	var value3 = ([]byte)("bvi_loop3")

	txn.Put(key1, value1)
	txn.Put(key2, value2)
	txn.Put(key3, value3)
	txn.Commit()
	Expect(client.checkIfExists(key1, value1)).To(BeTrue())
	Expect(client.checkIfExists(key2, value2)).To(BeTrue())
	Expect(client.checkIfExists(key3, value3)).To(BeTrue())
}

func TestDeleteInTxn(t *testing.T) {
	client := setupTest(t, true)
	defer client.teardownTest()

	txn := client.NewTxn()
	Expect(txn).ToNot(BeNil())

	var key1 = "/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop0"
	var value1 = ([]byte)("bvi_loop1")
	var key2 = "/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop1"
	var value2 = ([]byte)("bvi_loop2")
	var key3 = "/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop2"
	var value3 = ([]byte)("bvi_loop3")

	txn.Put(key1, value1)
	txn.Put(key2, value2)
	txn.Put(key3, value3)
	txn.Delete(key2)
	txn.Commit()
	Expect(client.checkIfExists(key1, value1)).To(BeTrue())
	Expect(client.checkIfExists(key2, value2)).To(BeFalse())
	Expect(client.checkIfExists(key3, value3)).To(BeTrue())
}

func TestListKeys(t *testing.T) {
	client := setupTest(t, true)
	defer client.teardownTest()

	var key1 = "/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop0"
	var value1 = ([]byte)("val 1")
	var key2 = "/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop1"
	var value2 = ([]byte)("val 2")
	var key3 = "/vnf-agent/agent_vpp_1/vpp/config/v1/bd/b1"
	var value3 = ([]byte)("val 3")
	var key4 = "/vnf-agent/agent_vpp_1/vpp/config/v1/bd/b2"
	var value4 = ([]byte)("val 4")
	var key5 = "/vnf-agent/agent_vpp_1/vpp/config/v2/bd/b1"
	var value5 = ([]byte)("val 5")

	txn := client.NewTxn()
	Expect(txn).ToNot(BeNil())
	txn.Put(key1, value1)
	txn.Put(key2, value2)
	txn.Put(key3, value3)
	txn.Put(key4, value4)
	txn.Put(key5, value5)
	txn.Commit()

	keys, err := client.ListKeys("/vnf-agent/agent_vpp_1/vpp/config/")
	Expect(err).ToNot(HaveOccurred())
	expectedKeys := []string{"v1/bd/b1", "v1/bd/b2", "v1/interface/bvi_loop0", "v1/interface/bvi_loop1", "v2/bd/b1"}
	for i := 0; i <= len(expectedKeys); i++ {
		k, _, all := keys.GetNext()
		if i == len(expectedKeys) {
			Expect(all).To(BeTrue())
			break
		}
		Expect(k).NotTo(BeNil())
		Expect(all).To(BeFalse())
		// verify that prefix is trimmed
		Expect(k).To(BeEquivalentTo(expectedKeys[i]))
	}
}

func TestListVals(t *testing.T) {
	client := setupTest(t, true)
	defer client.teardownTest()

	var key1 = "/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop0"
	var value1 = ([]byte)("val 1")
	var key2 = "/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop1"
	var value2 = ([]byte)("val 2")
	var key3 = "/vnf-agent/agent_vpp_1/vpp/config/v1/bd/b1"
	var value3 = ([]byte)("val 3")
	var key4 = "/vnf-agent/agent_vpp_1/vpp/config/v1/bd/b2"
	var value4 = ([]byte)("val 4")

	txn := client.NewTxn()
	Expect(txn).ToNot(BeNil())
	txn.Put(key1, value1)
	txn.Put(key2, value2)
	txn.Put(key3, value3)
	txn.Put(key4, value4)
	txn.Commit()

	keyVals, err := client.ListValues("/vnf-agent/agent_vpp_1/vpp/config/v1/")
	Expect(err).ToNot(HaveOccurred())
	expectedKeys := []string{"bd/b1", "bd/b2", "interface/bvi_loop0", "interface/bvi_loop1"}
	expectedValues := [][]byte{([]byte)("val 3"), ([]byte)("val 4"), ([]byte)("val 1"), ([]byte)("val 2")}
	for i := 0; i <= len(expectedKeys); i++ {
		kv, all := keyVals.GetNext()
		if i == len(expectedKeys) {
			Expect(all).To(BeTrue())
			break
		}
		Expect(kv).NotTo(BeNil())
		Expect(all).To(BeFalse())
		Expect(kv.GetValue()).To(BeEquivalentTo(expectedValues[i]))
		// verify that prefix is trimmed
		Expect(kv.GetKey()).To(BeEquivalentTo(expectedKeys[i]))
	}
}
