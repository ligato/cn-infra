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

package mux

import (
	"github.com/Shopify/sarama"
	"github.com/ligato/cn-infra/messaging"
	"github.com/onsi/gomega"
	"testing"
)

func TestMultiplexer(t *testing.T) {
	gomega.RegisterTestingT(t)
	mux, _, _ := GetMultiplexerMock(t)
	gomega.Expect(mux).NotTo(gomega.BeNil())

	c1 := mux.NewConnection("c1")
	gomega.Expect(c1).NotTo(gomega.BeNil())
	c2 := mux.NewConnection("c2")
	gomega.Expect(c2).NotTo(gomega.BeNil())

	ch1 := make(chan messaging.BytesMessage)
	ch2 := make(chan messaging.BytesMessage)

	err := c1.ConsumeTopic(ToBytesMsgChan(ch1), "topic1")
	gomega.Expect(err).To(gomega.BeNil())
	err = c2.ConsumeTopic(ToBytesMsgChan(ch2), "topic2", "topic3")
	gomega.Expect(err).To(gomega.BeNil())

	mux.Start()
	gomega.Expect(mux.started).To(gomega.BeTrue())

	// once the multiplexer is start an attempt to subscribe returns an error
	err = c1.ConsumeTopic(ToBytesMsgChan(ch1), "anotherTopic1")
	gomega.Expect(err).NotTo(gomega.BeNil())

	mux.Close()
	close(ch1)
	close(ch2)

}

func TestStopConsuming(t *testing.T) {
	gomega.RegisterTestingT(t)
	mux, _, _ := GetMultiplexerMock(t)
	gomega.Expect(mux).NotTo(gomega.BeNil())

	c1 := mux.NewConnection("c1")
	gomega.Expect(c1).NotTo(gomega.BeNil())
	c2 := mux.NewConnection("c2")
	gomega.Expect(c2).NotTo(gomega.BeNil())

	ch1 := make(chan messaging.BytesMessage)
	ch2 := make(chan messaging.BytesMessage)

	err := c1.ConsumeTopic(ToBytesMsgChan(ch1), "topic1")
	gomega.Expect(err).To(gomega.BeNil())
	err = c2.ConsumeTopic(ToBytesMsgChan(ch2), "topic2", "topic3")
	gomega.Expect(err).To(gomega.BeNil())

	mux.Start()
	gomega.Expect(mux.started).To(gomega.BeTrue())

	err = c1.StopConsuming("topic1")
	gomega.Expect(err).To(gomega.BeNil())

	// topic is not consumed
	err = c1.StopConsuming("Unknown topic")
	gomega.Expect(err).NotTo(gomega.BeNil())

	// topic consumed by a different connection
	err = c1.StopConsuming("topic2")
	gomega.Expect(err).NotTo(gomega.BeNil())

	mux.Close()
	close(ch1)
	close(ch2)

}

func TestSendSync(t *testing.T) {
	gomega.RegisterTestingT(t)
	mux, _, syncP := GetMultiplexerMock(t)
	gomega.Expect(mux).NotTo(gomega.BeNil())

	c1 := mux.NewConnection("c1")
	gomega.Expect(c1).NotTo(gomega.BeNil())

	mux.Start()
	gomega.Expect(mux.started).To(gomega.BeTrue())

	syncP.ExpectSendMessageAndSucceed()
	_, err := c1.SendSyncByte("topic", []byte("key"), []byte("value"))
	gomega.Expect(err).To(gomega.BeNil())

	syncP.ExpectSendMessageAndSucceed()
	_, err = c1.SendSyncString("topic", "key", "value")
	gomega.Expect(err).To(gomega.BeNil())

	syncP.ExpectSendMessageAndSucceed()
	_, err = c1.SendSyncMessage("topic", sarama.ByteEncoder([]byte("key")), sarama.ByteEncoder([]byte("value")))
	gomega.Expect(err).To(gomega.BeNil())

	publisher := c1.NewSyncPublisher("test")
	syncP.ExpectSendMessageAndSucceed()
	publisher.Publish("key", []byte("val"))

	mux.Close()
}

func TestSendAsync(t *testing.T) {
	gomega.RegisterTestingT(t)
	mux, asyncP, _ := GetMultiplexerMock(t)
	gomega.Expect(mux).NotTo(gomega.BeNil())

	c1 := mux.NewConnection("c1")
	gomega.Expect(c1).NotTo(gomega.BeNil())

	mux.Start()
	gomega.Expect(mux.started).To(gomega.BeTrue())

	asyncP.ExpectInputAndSucceed()
	c1.SendAsyncByte("topic", []byte("key"), []byte("value"), nil, nil, nil)

	asyncP.ExpectInputAndSucceed()
	c1.SendAsyncString("topic", "key", "value", nil, nil, nil)

	asyncP.ExpectInputAndSucceed()
	c1.SendAsyncMessage("topic", sarama.ByteEncoder([]byte("key")), sarama.ByteEncoder([]byte("value")), nil, nil, nil)

	publisher := c1.NewAsyncPublisher("test", nil, nil)
	asyncP.ExpectInputAndSucceed()
	publisher.Publish("key", []byte("val"))

	mux.Close()
}
