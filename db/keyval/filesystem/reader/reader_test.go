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

package reader_test

import (
	"testing"

	"github.com/ligato/cn-infra/db/keyval/filesystem/reader"

	. "github.com/onsi/gomega"
)

func TestCompareFiles(t *testing.T) {
	RegisterTestingT(t)

	// Original data
	origin := make(map[string][]byte)
	origin["key1"] = []byte("data1")
	origin["key2"] = []byte("data2")
	origin["key3"] = []byte("data3")
	origin["key4"] = []byte("data4")
	origin["key5"] = []byte("data5")
	origin["key6"] = []byte("data6")

	// New data
	var change []reader.FileEntry
	// Unchanged
	change = append(change, reader.FileEntry{Key: "key1", Value: []byte("data1")})
	change = append(change, reader.FileEntry{Key: "key2", Value: []byte("data2")})
	// Changed
	change = append(change, reader.FileEntry{Key: "key3", Value: []byte("changedData1")})
	change = append(change, reader.FileEntry{Key: "key4", Value: []byte("changedData2")})
	// Added
	change = append(change, reader.FileEntry{Key: "key7", Value: []byte("newData1")})
	change = append(change, reader.FileEntry{Key: "key8", Value: []byte("newData2")})
	// key5 and key6 was removed

	changeData := &reader.File{Path: "path1", Data: change}

	changed, removed := changeData.CompareTo(origin)
	Expect(changed).To(HaveLen(4)) // Changed + Added
	for _, change := range changed {
		Expect([]string{"key3", "key4", "key7", "key8"}).To(ContainElement(change.Key))
		switch change.Key {
		case "key3":
			Expect(change.Value).To(BeEquivalentTo([]byte("changedData1")))
		case "key4":
			Expect(change.Value).To(BeEquivalentTo([]byte("changedData2")))
		case "key7":
			Expect(change.Value).To(BeEquivalentTo([]byte("newData1")))
		case "key8":
			Expect(change.Value).To(BeEquivalentTo([]byte("newData2")))
		}
	}
	Expect(removed).To(HaveLen(2)) // Delete
	for _, remove := range removed {
		Expect([]string{"key5", "key6"}).To(ContainElement(remove.Key))
		switch remove.Key {
		case "key5":
			Expect(remove.Value).To(BeEquivalentTo([]byte("data5")))
		case "key6":
			Expect(remove.Value).To(BeEquivalentTo([]byte("data6")))
		}
	}
}
