package mux

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/messaging"
	"github.com/ligato/cn-infra/messaging/kafka/client"
)

// ProtoConnection is an entity that provides access to shared producers/consumers of multiplexer.
// The value of message are marshaled and unmarshaled to/from proto.message behind the scene.
type ProtoConnection struct {
	// multiplexer is used for access to kafka brokers
	multiplexer *Multiplexer

	// name identifies the connection
	name string

	// serializer marshals and unmarshals data to/from proto.Message
	serializer keyval.Serializer
}

type protoSyncPublisherKafka struct {
	conn  *ProtoConnection
	topic string
}

type protoAsyncPublisherKafka struct {
	conn       *ProtoConnection
	topic      string
	successClb func(messaging.ProtoMessage)
	errClb     func(messaging.ProtoMessageErr)
}

// SendSyncMessage sends a message using the sync API
func (conn *ProtoConnection) SendSyncMessage(topic string, key string, value proto.Message) (offset int64, err error) {
	data, err := conn.serializer.Marshal(value)
	if err != nil {
		return 0, err
	}
	msg, err := conn.multiplexer.syncProducer.SendMsg(topic, sarama.StringEncoder(key), sarama.ByteEncoder(data))
	if err != nil {
		return 0, err
	}
	return msg.Offset, err
}

// SendAsyncMessage sends a message using the async API
func (conn *ProtoConnection) SendAsyncMessage(topic string, key string, value proto.Message, meta interface{}, successClb func(messaging.ProtoMessage), errClb func(messaging.ProtoMessageErr)) error {
	data, err := conn.serializer.Marshal(value)
	if err != nil {
		return err
	}
	succByteClb := func(msg messaging.BytesMessage) {
		protoMsg := &client.ProtoProducerMessage{
			ProducerMessage: msg.(*client.ProducerMessage),
			Serializer:      conn.serializer,
		}
		successClb(protoMsg)
	}

	errByteClb := func(msg messaging.BytesMessageErr) {
		kafkaMsg := msg.(*client.ProducerError)
		protoMsg := &client.ProtoProducerMessageErr{
			ProtoProducerMessage: &client.ProtoProducerMessage{
				ProducerMessage: kafkaMsg.ProducerMessage,
				Serializer:      conn.serializer,
			},
			Err: kafkaMsg.Err,
		}
		errClb(protoMsg)
	}

	auxMeta := &asyncMeta{successClb: succByteClb, errorClb: errByteClb, usersMeta: meta}
	conn.multiplexer.asyncProducer.SendMsg(topic, sarama.StringEncoder(key), sarama.ByteEncoder(data), auxMeta)
	return nil
}

// ConsumeTopic is called to start consuming given topics.
// Function can be called until the multiplexer is started, it returns an error otherwise.
// The provided channel should be buffered, otherwise messages might be lost.
func (conn *ProtoConnection) ConsumeTopic(msgClb func(messaging.ProtoMessage), topics ...string) error {
	conn.multiplexer.rwlock.Lock()
	defer conn.multiplexer.rwlock.Unlock()

	if conn.multiplexer.started {
		return fmt.Errorf("ConsumeTopic can be called only if the multiplexer has not been started yet")
	}

	byteClb := func(bm messaging.BytesMessage) {
		pm := client.NewProtoConsumerMessage(bm.(*client.ConsumerMessage), conn.serializer)
		msgClb(pm)
	}

	for _, topic := range topics {
		// check if we have already consumed the topic and partition
		subs, found := conn.multiplexer.mapping[topic]

		if !found {
			subs = &map[string]func(messaging.BytesMessage){}
			conn.multiplexer.mapping[topic] = subs
		}
		// add subscription to consumerList
		(*subs)[conn.name] = byteClb
		conn.multiplexer.mapping[topic] = subs
	}
	return nil
}

// StopConsuming cancels the previously created subscription for consuming the topic.
func (conn *ProtoConnection) StopConsuming(topic string) error {
	return conn.multiplexer.stopConsuming(topic, conn.name)
}

// NewSyncPublisher creates a new instance of protoSyncPublisherKafka that allows to publish sync kafka messages using common messaging API
func (conn *ProtoConnection) NewSyncPublisher(topic string) messaging.ProtoPublisher {
	return &protoSyncPublisherKafka{conn, topic}
}

// Publish publishes a message into kafka
func (p *protoSyncPublisherKafka) Publish(key string, message proto.Message) error {
	_, err := p.conn.SendSyncMessage(p.topic, key, message)
	return err
}

// NewAsyncPublisher creates a new instance of protoAsyncPublisherKafka that allows to publish sync kafka messages using common messaging API
func (conn *ProtoConnection) NewAsyncPublisher(topic string, successClb func(messaging.ProtoMessage), errorClb func(messaging.ProtoMessageErr)) messaging.ProtoPublisher {
	return &protoAsyncPublisherKafka{conn, topic, successClb, errorClb}
}

// Publish publishes a message into kafka
func (p *protoAsyncPublisherKafka) Publish(key string, message proto.Message) error {
	return p.conn.SendAsyncMessage(p.topic, key, message, nil, p.successClb, p.errClb)
}
