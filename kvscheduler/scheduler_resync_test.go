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

package kvscheduler

import (
	"context"
	"strings"
	"testing"
	"time"
	. "github.com/onsi/gomega"

	"github.com/ligato/cn-infra/kvscheduler/test"
	. "github.com/ligato/cn-infra/kvscheduler/api"
)

func TestEmptyResync(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin()
	err := scheduler.Init()
	Expect(err).To(BeNil())

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	descriptor1 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:            descriptor1Name,
		KeySelector:     prefixSelector(prefixA),
		NBKeyPrefixes:   []string{prefixA},
		WithMetadata:    true,
		DumpIsSupported: true,
	}, mockSB)

	// register descriptor with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)

	// get metadata map created for the descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.GetName())
	_, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// transaction history should be initially empty
	Expect(scheduler.getTransactionHistory(time.Now())).To(BeEmpty())

	// run transaction with empty resync
	startTime := time.Now()
	kvErrors, txnError := scheduler.StartNBTransaction().Resync([]KeyValueDataPair{}).Commit(context.Background())
	stopTime := time.Now()
	Expect(txnError).ShouldNot(HaveOccurred())
	Expect(kvErrors).To(BeEmpty())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	Expect(mockSB.GetValues(nil)).To(BeEmpty())

	// check metadata
	Expect(metadataMap.ListAllNames()).To(BeEmpty())

	// check executed operations
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(1))
	Expect(opHistory[0].OpType).To(Equal(test.Dump))
	Expect(opHistory[0].CorrelateDump).To(BeEmpty())

	// single transaction consisted of zero operations
	txnHistory := scheduler.getTransactionHistory(time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn := txnHistory[0]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(startTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(stopTime)).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(0))
	Expect(txn.txnType).To(BeEquivalentTo(nbTransaction))
	Expect(txn.isResync).To(BeTrue())
	Expect(txn.values).To(BeEmpty())
	Expect(txn.preErrors).To(BeEmpty())
	Expect(txn.planned).To(BeEmpty())
	Expect(txn.executed).To(BeEmpty())

	// close scheduler
	err = scheduler.Close()
	Expect(err).To(BeNil())
}

func TestResyncWithEmptySB(t *testing.T) {
	RegisterTestingT(t)

	// prepare KV Scheduler
	scheduler := NewPlugin()
	err := scheduler.Init()
	Expect(err).To(BeNil())

	// prepare mocks
	mockSB := test.NewMockSouthbound()
	descriptor1 := test.NewMockDescriptor(&test.MockDescriptorArgs{
		Name:            descriptor1Name,
		KeySelector:     prefixSelector(prefixA),
		NBKeyPrefixes:   []string{prefixA},
		ValueBuilder:    func(key string, valueData interface{}) (value Value, err error) {
			label := strings.TrimPrefix(key, prefixA)
			items, ok := valueData.([]string)
			if !ok {
				return nil, ErrInvalidValueDataType(key)
			}
			return test.NewArrayValue(Object, label, items...), nil
		},
		DependencyBuilder: func(key string, value Value) []Dependency {
			if key == prefixA + baseValue2 {
				depKey := prefixA + baseValue1 + "/item1" // base value depends on a derived value
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			if key == prefixA + baseValue1 + "/item2" {
				depKey := prefixA + baseValue2 + "/item1" // derived value depends on another derived value
				return []Dependency{
					{Label: depKey, Key: depKey},
				}
			}
			return nil
		},
		DerValuesBuilder: test.ArrayValueDerBuilder,
		WithMetadata:    true,
		DumpIsSupported: true,
	}, mockSB)

	// register descriptor with the scheduler
	scheduler.RegisterKVDescriptor(descriptor1)

	// get metadata map created for the descriptor
	metadataMap := scheduler.GetMetadataMap(descriptor1.GetName())
	nameToInteger, withMetadataMap := metadataMap.(test.NameToInteger)
	Expect(withMetadataMap).To(BeTrue())

	// run resync transaction with empty SB
	startTime := time.Now()
	values := []KeyValueDataPair{
		{Key: prefixA + baseValue2, ValueData: []string{"item1"}},
		{Key: prefixA + baseValue1, ValueData: []string{"item1", "item2"}},
	}
	kvErrors, txnError := scheduler.StartNBTransaction().Resync(values).Commit(context.Background())
	stopTime := time.Now()
	Expect(txnError).ShouldNot(HaveOccurred())
	Expect(kvErrors).To(BeEmpty())

	// check the state of SB
	Expect(mockSB.GetKeysWithInvalidData()).To(BeEmpty())
	// -> base value 1
	value := mockSB.GetValue(prefixA + baseValue1)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue1, "item1", "item2"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(*test.OnlyInteger).Integer).To(BeEquivalentTo(0))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 1
	value = mockSB.GetValue(prefixA + baseValue1 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item1", "item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item2 derived from base value 1
	value = mockSB.GetValue(prefixA + baseValue1 + "/item2")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item2", "item2"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> base value 2
	value = mockSB.GetValue(prefixA + baseValue2)
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewArrayValue(Object, baseValue2, "item1"))).To(BeTrue())
	Expect(value.Metadata).ToNot(BeNil())
	Expect(value.Metadata.(*test.OnlyInteger).Integer).To(BeEquivalentTo(1))
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	// -> item1 derived from base value 2
	value = mockSB.GetValue(prefixA + baseValue2 + "/item1")
	Expect(value).ToNot(BeNil())
	Expect(value.Value.Equivalent(test.NewStringValue(Object, "item1", "item1"))).To(BeTrue())
	Expect(value.Metadata).To(BeNil())
	Expect(value.Origin).To(BeEquivalentTo(FromNB))
	Expect(mockSB.GetValues(nil)).To(HaveLen(5))

	// check metadata
	metadata, exists := nameToInteger.LookupByName(baseValue1)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(0))
	metadata, exists = nameToInteger.LookupByName(baseValue2)
	Expect(exists).To(BeTrue())
	Expect(metadata.GetInteger()).To(BeEquivalentTo(1))

	// check executed operations
	opHistory := mockSB.PopHistoryOfOps()
	Expect(opHistory).To(HaveLen(6))
	operation := opHistory[0]
	Expect(operation.OpType).To(Equal(test.Dump))
	checkValuesForCorrelation(operation.CorrelateDump, []KVWithMetadata{
		{
			Key: prefixA + baseValue1,
			Value: test.NewArrayValue(Object, baseValue1, "item1", "item2"),
			Metadata: nil,
			Origin: FromNB,
		},
		{
			Key: prefixA + baseValue2,
			Value: test.NewArrayValue(Object, baseValue2, "item1"),
			Metadata: nil,
			Origin: FromNB,
		},
	})
	operation = opHistory[1]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[2]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[3]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[4]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue2 + "/item1"))
	Expect(operation.Err).To(BeNil())
	operation = opHistory[5]
	Expect(operation.OpType).To(Equal(test.Add))
	Expect(operation.Descriptor).To(BeEquivalentTo(descriptor1Name))
	Expect(operation.Key).To(BeEquivalentTo(prefixA + baseValue1 + "/item2"))
	Expect(operation.Err).To(BeNil())

	// single transaction consisted of 6 operations
	txnHistory := scheduler.getTransactionHistory(time.Now())
	Expect(txnHistory).To(HaveLen(1))
	txn := txnHistory[0]
	Expect(txn.preRecord).To(BeFalse())
	Expect(txn.start.After(startTime)).To(BeTrue())
	Expect(txn.start.Before(txn.stop)).To(BeTrue())
	Expect(txn.stop.Before(stopTime)).To(BeTrue())
	Expect(txn.seqNum).To(BeEquivalentTo(0))
	Expect(txn.txnType).To(BeEquivalentTo(nbTransaction))
	Expect(txn.isResync).To(BeTrue())
	checkRecordedValues(txn.values, []recordedKVPair{
		{key: prefixA + baseValue1, value: &recordedValue{valueType: Object, label: baseValue1, string: "[item1,item2]"}},
		{key: prefixA + baseValue2, value: &recordedValue{valueType: Object, label: baseValue2, string: "[item1]"}},
	})
	Expect(txn.preErrors).To(BeEmpty())

	txnOps := recordedTxnOps{
		{
			operation:  add,
			key:        prefixA + baseValue1,
			newValue:   &recordedValue{valueType: Object, label: baseValue1, string: "[item1,item2]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item1",
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item2",
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			isPending:  true,
		},
		{
			operation:  add,
			key:        prefixA + baseValue2,
			newValue:   &recordedValue{valueType: Object, label: baseValue2, string: "[item1]"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue2 + "/item1",
			newValue:   &recordedValue{valueType: Object, label: "item1", string: "item1"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
		},
		{
			operation:  add,
			key:        prefixA + baseValue1 + "/item2",
			prevValue:  &recordedValue{valueType: Object, label: "item2", string: "item2"},
			newValue:   &recordedValue{valueType: Object, label: "item2", string: "item2"},
			prevOrigin: FromNB,
			newOrigin:  FromNB,
			wasPending: true,
		},
	}
	checkTxnOperations(txn.planned, txnOps)
	checkTxnOperations(txn.executed, txnOps)

	// close scheduler
	err = scheduler.Close()
	Expect(err).To(BeNil())
}

