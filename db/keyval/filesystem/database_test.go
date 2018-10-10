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

package filesystem_test

import (
	"testing"

	"github.com/ligato/cn-infra/db/keyval/filesystem"
	. "github.com/onsi/gomega"
)

const (
	// Files
	file1 = "/path/to/file1"
	file2 = "/path/to/file2"
	file3 = "/path/to/file3"
	// Keys
	ifKey1  = "/vpp/config/v1/interfaces/if1"
	ifKey2  = "/vpp/config/v1/interfaces/if2"
	bdKey1  = "/vpp/config/v1/bd/bd1"
	bdKey2  = "/vpp/config/v1/bd/bd2"
	fibKey1 = "/vpp/config/v1/fib/fib1"
	fibKey2 = "/vpp/config/v1/fib/fib2"
)

func TestAddDelEntry(t *testing.T) {
	RegisterTestingT(t)

	// Test Add
	db := filesystem.NewDbClient()
	db.Add(file1, ifKey1, []byte(ifKey1))
	db.Add(file1, ifKey2, []byte(ifKey2))
	db.Add(file2, bdKey1, []byte(bdKey1))

	dataMap := db.GetDataFromFile(file1)
	Expect(dataMap).To(HaveLen(2))
	Expect(dataMap[ifKey1]).To(BeEquivalentTo([]byte(ifKey1)))
	Expect(dataMap[ifKey2]).To(BeEquivalentTo([]byte(ifKey2)))

	dataMap = db.GetDataFromFile(file2)
	Expect(dataMap).To(HaveLen(1))
	Expect(dataMap[bdKey1]).To(BeEquivalentTo([]byte(bdKey1)))

	// Test Delete
	db.Delete(file1, ifKey1)
	db.Delete(file2, bdKey1)

	dataMap = db.GetDataFromFile(file1)
	Expect(dataMap).To(HaveLen(1))
	Expect(dataMap[ifKey2]).To(BeEquivalentTo([]byte(ifKey2)))

	dataMap = db.GetDataFromFile(file2)
	Expect(dataMap).To(HaveLen(0))
}

func TestModifyEntry(t *testing.T) {
	RegisterTestingT(t)

	db := filesystem.NewDbClient()
	db.Add(file1, bdKey1, []byte(bdKey1))

	dataMap := db.GetDataFromFile(file1)
	Expect(dataMap).To(HaveLen(1))
	Expect(dataMap[bdKey1]).To(BeEquivalentTo([]byte(bdKey1)))

	// Modify value
	db.Add(file1, bdKey1, []byte(bdKey2))

	dataMap = db.GetDataFromFile(file1)
	Expect(dataMap).To(HaveLen(1))
	Expect(dataMap[bdKey1]).To(BeEquivalentTo([]byte(bdKey2)))

	// Move to different key
	db.Delete(file1, bdKey1)
	db.Add(file1, bdKey2, []byte(bdKey2))

	dataMap = db.GetDataFromFile(file1)
	Expect(dataMap).To(HaveLen(1))
	Expect(dataMap[bdKey2]).To(BeEquivalentTo([]byte(bdKey2)))
}

func TestDeleteNonExisting(t *testing.T) {
	RegisterTestingT(t)

	db := filesystem.NewDbClient()
	db.Add(file1, ifKey1, []byte(ifKey1))
	db.Delete(file1, ifKey2)

	dataMap := db.GetDataFromFile(file1)
	Expect(dataMap).To(HaveLen(1))

	db.Delete(file1, ifKey1)

	dataMap = db.GetDataFromFile(file1)
	Expect(dataMap).To(HaveLen(0))

	db.Delete(file2, ifKey1)

	dataMap = db.GetDataFromFile(file2)
	Expect(dataMap).To(HaveLen(0))
}

func TestDeleteFile(t *testing.T) {
	RegisterTestingT(t)

	// Test Add
	db := filesystem.NewDbClient()
	db.Add(file1, ifKey1, []byte(ifKey1))
	db.Add(file1, ifKey2, []byte(ifKey2))
	db.Add(file3, fibKey1, []byte(fibKey1))
	db.Add(file3, fibKey2, []byte(fibKey2))

	dataMap := db.GetDataFromFile(file1)
	Expect(dataMap).To(HaveLen(2))

	dataMap = db.GetDataFromFile(file3)
	Expect(dataMap).To(HaveLen(2))

	// Remove first file
	db.DeleteFile(file1)

	dataMap = db.GetDataFromFile(file1)
	Expect(dataMap).To(HaveLen(0))

	dataMap = db.GetDataFromFile(file3)
	Expect(dataMap).To(HaveLen(2))

	// Remove second file
	db.DeleteFile(file3)

	dataMap = db.GetDataFromFile(file1)
	Expect(dataMap).To(HaveLen(0))

	dataMap = db.GetDataFromFile(file3)
	Expect(dataMap).To(HaveLen(0))
}

func TestGetKeysFromPrefix(t *testing.T) {
	RegisterTestingT(t)

	db := filesystem.NewDbClient()
	db.Add(file1, ifKey1, []byte(ifKey1))
	db.Add(file1, ifKey2, []byte(ifKey2))
	db.Add(file1, fibKey1, []byte(fibKey1))
	db.Add(file2, fibKey2, []byte(fibKey2))

	// From the same file
	keys := db.GetKeysForPrefix("/vpp/config/v1/interfaces")
	Expect(keys).To(HaveLen(2))
	Expect(keys).To(ConsistOf(ifKey1, ifKey2))

	// From different files
	keys = db.GetKeysForPrefix("/vpp/config/v1/fib")
	Expect(keys).To(HaveLen(2))
	Expect(keys).To(ConsistOf(fibKey1, fibKey2))
}

func TestGetValuesFromPrefix(t *testing.T) {
	RegisterTestingT(t)

	db := filesystem.NewDbClient()
	db.Add(file1, ifKey1, []byte(ifKey1))
	db.Add(file1, ifKey2, []byte(ifKey2))
	db.Add(file1, fibKey1, []byte(fibKey1))
	db.Add(file2, fibKey2, []byte(fibKey2))

	// From the same file
	keys := db.GetValuesForPrefix("/vpp/config/v1/interfaces")
	Expect(keys).To(HaveLen(2))
	Expect(keys).To(ConsistOf([]byte(ifKey1), []byte(ifKey2)))

	// From different files
	keys = db.GetValuesForPrefix("/vpp/config/v1/fib")
	Expect(keys).To(HaveLen(2))
	Expect(keys).To(ConsistOf([]byte(fibKey1), []byte(fibKey2)))
}

func TestGetDataForKey(t *testing.T) {
	RegisterTestingT(t)

	db := filesystem.NewDbClient()
	db.Add(file1, ifKey1, []byte(ifKey1))
	db.Add(file1, bdKey1, []byte(bdKey1))
	db.Add(file1, fibKey1, []byte(fibKey1))

	// Existing
	data, ok := db.GetDataForKey(ifKey1)
	Expect(ok).To(BeTrue())
	Expect(data).To(BeEquivalentTo([]byte(ifKey1)))

	data, ok = db.GetDataForKey(bdKey1)
	Expect(ok).To(BeTrue())
	Expect(data).To(BeEquivalentTo([]byte(bdKey1)))

	data, ok = db.GetDataForKey(fibKey1)
	Expect(ok).To(BeTrue())
	Expect(data).To(BeEquivalentTo([]byte(fibKey1)))

	// Non-existing
	data, ok = db.GetDataForKey(ifKey2)
	Expect(ok).To(BeFalse())
	Expect(data).To(BeNil())

	data, ok = db.GetDataForKey(bdKey2)
	Expect(ok).To(BeFalse())
	Expect(data).To(BeNil())
}

func TestGetDataForPathAndKey(t *testing.T) {
	RegisterTestingT(t)

	db := filesystem.NewDbClient()
	db.Add(file1, ifKey1, []byte(ifKey1))
	db.Add(file2, bdKey1, []byte(bdKey1))

	// Existing
	data, ok := db.GetDataForPathAndKey(file1, ifKey1)
	Expect(ok).To(BeTrue())
	Expect(data).To(BeEquivalentTo([]byte(ifKey1)))

	data, ok = db.GetDataForPathAndKey(file2, bdKey1)
	Expect(ok).To(BeTrue())
	Expect(data).To(BeEquivalentTo([]byte(bdKey1)))

	// Non-existing path
	data, ok = db.GetDataForPathAndKey(file2, ifKey1)
	Expect(ok).To(BeFalse())
	Expect(data).To(BeNil())

	data, ok = db.GetDataForPathAndKey(file3, bdKey1)
	Expect(ok).To(BeFalse())
	Expect(data).To(BeNil())

	// Non-existing key
	data, ok = db.GetDataForPathAndKey(file1, ifKey2)
	Expect(ok).To(BeFalse())
	Expect(data).To(BeNil())

	data, ok = db.GetDataForPathAndKey(file2, bdKey2)
	Expect(ok).To(BeFalse())
	Expect(data).To(BeNil())
}
