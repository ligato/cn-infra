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

package messaging

import (
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/db/keyval"
)

// Mux defines API for the plugins that use access to kafka brokers.
type Mux interface {
	NewConnection(name string) Connection
	NewProtoConnection(name string) ProtoConnection
}

// Encoder defines an interface that is used as argument of producer functions.
// It wraps the sarama.Encoder
type Encoder interface {
	sarama.Encoder
}

type Connection interface {
	ConsumeTopic(msgClb func(BytesMessage), topics ...string) error
	StopConsuming(topic string) error
	SendSyncByte(topic string, key []byte, value []byte) (offset int64, err error)
	SendSyncString(topic string, key string, value string) (offset int64, err error)
	SendSyncMessage(topic string, key Encoder, value Encoder) (offset int64, err error)
	SendAsyncByte(topic string, key []byte, value []byte, meta interface{}, successClb func(BytesMessage), errClb func(err BytesMessageErr))
	SendAsyncString(topic string, key string, value string, meta interface{}, successClb func(BytesMessage), errClb func(err BytesMessageErr))
	SendAsyncMessage(topic string, key Encoder, value Encoder, meta interface{}, successClb func(BytesMessage), errClb func(err BytesMessageErr))
	NewSyncPublisher(topic string) BytesPublisher
	NewAsyncPublisher(topic string, successClb func(BytesMessage), errorClb func(err BytesMessageErr)) BytesPublisher
}

type ProtoConnection interface {
	ConsumeTopic(msgChan func(ProtoMessage), topics ...string) error
	StopConsuming(topic string) error
	SendSyncMessage(topic string, key string, value proto.Message) (offset int64, err error)
	SendAsyncMessage(topic string, key string, value proto.Message, meta interface{}, successChan func(ProtoMessage), errorClb func(err ProtoMessageErr)) error
	NewSyncPublisher(topic string) ProtoPublisher
	NewAsyncPublisher(topic string, successClb func(ProtoMessage), errorClb func(err ProtoMessageErr)) ProtoPublisher
}

// BytesPublisher allows to publish a message of type []bytes into messaging system.
type BytesPublisher interface {
	Publish(key string, data []byte) error
}

// ProtoPublisher allows to publish a message of type proto.Message into messaging system.
type ProtoPublisher interface {
	Publish(key string, data proto.Message) error
}

// BytesMessage defines functions for inspection of a message received from messaging system.
type BytesMessage interface {
	keyval.BytesKvPair
	GetTopic() string
}

// BytesMessageErr defines functions for inspection of a message that kafka was unable to publish successfully.
type BytesMessageErr interface {
	BytesMessage
	Error() error
}

// ProtoMessage defines functions for inspection of a message receive from messaging system.
type ProtoMessage interface {
	keyval.ProtoKvPair
	GetTopic() string
}

type ProtoMessageErr interface {
	ProtoMessage
	Error() error
}
