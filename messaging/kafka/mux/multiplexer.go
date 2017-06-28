package mux

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/db/keyval"
	lg "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/messaging/kafka/client"
	"sync"
	"time"
)

// Multiplexer encapsulates clients to kafka cluster (syncProducer, asyncProducer, consumer).
// It allows to create multiple Connections that use multiplexer's clients for communication
// with kafka cluster. The aim of Multiplexer is to decrease the number of connections needed.
// The set of topics to be consumed by Connections needs to be selected before the underlying
// consumer in Multiplexer is started. Once the Multiplexer's consumer has been
// started new topics can not be added.
type Multiplexer struct {
	// consumer used by the Multiplexer
	consumer *client.Consumer
	// syncProducer used by the Multiplexer
	syncProducer *client.SyncProducer
	// asyncProducer used by the Multiplexer
	asyncProducer *client.AsyncProducer

	// name is used for identification of stored last consumed offset in kafka. This allows
	// to follow up messages after restart.
	name string

	// guards access to mapping and started flag
	rwlock sync.RWMutex

	// started denotes whether the multiplexer is dispatching the messages or accepting subscriptions to
	// consume a topic. Once the multiplexer is started, new subscription can not be added.
	started bool

	// Mapping provides the mapping of subscribed consumers organized by topics(key of the first map)
	// name of the consumer(key of the second map)
	mapping map[string]*map[string]chan *client.ConsumerMessage

	// factory that crates consumer used in the Multiplexer
	consumerFactory func(topics []string, groupId string) (*client.Consumer, error)
}

// Connection is an entity that provides access to shared producers/consumers of multiplexer.
type Connection struct {
	// multiplexer is used for access to kafka brokers
	multiplexer *Multiplexer

	// name identifies the connection
	name string
}

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

// asyncMeta is auxiliary structure used by Multiplexer to distribute consumer messages
type asyncMeta struct {
	successChan chan *client.ProducerMessage
	errorChan   chan *client.ProducerError
	usersMeta   interface{}
}

// NewMultiplexer creates new instance of Kafka Multiplexer
func NewMultiplexer(consumerFactory ConsumerFactory, syncP *client.SyncProducer, asyncP *client.AsyncProducer, name string) *Multiplexer {
	cl := &Multiplexer{consumerFactory: consumerFactory,
		syncProducer:  syncP,
		asyncProducer: asyncP,
		name:          name,
		mapping:       map[string]*map[string]chan *client.ConsumerMessage{}}

	go cl.watchAsyncProducerChannels()
	return cl
}

func (mux *Multiplexer) watchAsyncProducerChannels() {
	for {
		select {
		case err := <-mux.asyncProducer.Config.ErrorChan:
			log.Println("Failed to produce message", err.Err)
			errMsg := err.Msg

			if errMeta, ok := errMsg.Metadata.(*asyncMeta); ok && errMeta.errorChan != nil {
				err.Msg.Metadata = errMeta.usersMeta
				select {
				case errMeta.errorChan <- err:
				default:
					//case <-time.NewTimer(time.Second).C:
					log.Warn("Unable to send error notification")
				}
			}
		case success := <-mux.asyncProducer.Config.SuccessChan:

			if succMeta, ok := success.Metadata.(*asyncMeta); ok && succMeta.successChan != nil {
				success.Metadata = succMeta.usersMeta
				select {
				case succMeta.successChan <- success:
				default:
					//case <-time.NewTimer(time.Second).C:
					log.Warn("Unable to send success notification")
				}
			}
		case <-mux.asyncProducer.GetCloseChannel():
			log.Debug("Closing watch loop for async producer")
		}
	}
}

// Start should be called once all the Connections have been subscribed
// for topic consumption. An attempt to start consuming a topic after the multiplexer is started
// returns an error.
func (mux *Multiplexer) Start() error {
	mux.rwlock.Lock()
	defer mux.rwlock.Unlock()
	var err error

	if mux.started {
		return fmt.Errorf("Multiplexer has been started already")
	}

	// block further consumer consumers
	mux.started = true

	var topics []string

	for topic := range mux.mapping {
		topics = append(topics, topic)
	}

	log.WithFields(lg.Fields{"topics": topics}).Debug("Consuming started")

	mux.consumer, err = mux.consumerFactory(topics, mux.name)
	if err != nil {
		log.Error(err)
		return err
	}

	go mux.genericConsumer()

	return nil
}

// Close cleans up the resources used by the Multiplexer
func (mux *Multiplexer) Close() {
	mux.consumer.Close()
	mux.syncProducer.Close()
	mux.asyncProducer.Close()
}

// NewConnection creates instance of the Connection that will be provide access to shared Multiplexer's clients.
func (mux *Multiplexer) NewConnection(name string) *Connection {
	return &Connection{multiplexer: mux, name: name}
}

// NewProtoConnection creates instance of the ProtoConnection that will be provide access to shared Multiplexer's clients.
func (mux *Multiplexer) NewProtoConnection(name string, serializer keyval.Serializer) *ProtoConnection {
	return &ProtoConnection{multiplexer: mux, serializer: serializer, name: name}
}

func (mux *Multiplexer) propagateMessage(msg *client.ConsumerMessage) {
	mux.rwlock.RLock()
	defer mux.rwlock.RUnlock()

	if msg == nil {
		return
	}
	cons, found := mux.mapping[msg.Topic]

	// notify consumers
	if found {
		for _, ch := range *cons {
			// if we are not able to write into the channel we should skip the receiver
			// and report an error to avoid deadlock
			log.Debug("offset ", msg.Offset, string(msg.Value), string(msg.Key), msg.Partition)

			select {
			case ch <- msg:
			case <-time.After(time.Second):
				log.Error("Unable to deliver message before the timeout.")
			}
		}
	}
}

// genericConsumer handles incoming messages to the multiplexer and distributes them among the subscribers
func (mux *Multiplexer) genericConsumer() {
	log.Debug("Generic consumer started")
	for {
		select {
		case <-mux.consumer.GetCloseChannel():
			log.Debug("Closing consumer")
			return
		case msg := <-mux.consumer.Config.RecvMessageChan:
			log.Debug("Kafka message received")
			mux.propagateMessage(msg)
			// Mark offset as read. If the Multiplexer is restarted it
			// continues to receive message after the last committed offset.
			mux.consumer.MarkOffset(msg, "")
		case err := <-mux.consumer.Config.RecvErrorChan:
			log.Error("Received partitionConsumer error ", err)
		}
	}

}

func (mux *Multiplexer) stopConsuming(topic string, name string) error {
	mux.rwlock.Lock()
	defer mux.rwlock.Unlock()

	subs, found := mux.mapping[topic]
	if !found {
		return fmt.Errorf("Topic %s was not consumed by '%s'", topic, name)
	}
	_, found = (*subs)[name]
	if !found {
		return fmt.Errorf("Topic %s was not consumed by '%s'", topic, name)
	}
	delete(*subs, name)
	return nil
}

// ConsumeTopic is called to start consuming of a topic.
// Function can be called until the multiplexer is started, it returns an error otherwise.
// The provided channel should be buffered, otherwise messages might be lost.
func (conn *Connection) ConsumeTopic(msgChan chan *client.ConsumerMessage, topics ...string) error {
	conn.multiplexer.rwlock.Lock()
	defer conn.multiplexer.rwlock.Unlock()

	if conn.multiplexer.started {
		return fmt.Errorf("ConsumeTopic can be called only if the multiplexer has not been started yet")
	}

	for _, topic := range topics {
		// check if we have already consumed the topic and partition
		subs, found := conn.multiplexer.mapping[topic]

		if !found {
			subs = &map[string]chan *client.ConsumerMessage{}
			conn.multiplexer.mapping[topic] = subs
		}
		// add subscription to consumerList
		(*subs)[conn.name] = msgChan
		conn.multiplexer.mapping[topic] = subs
	}
	return nil
}

// StopConsuming cancels the previously created subscription for consuming the topic.
func (conn *Connection) StopConsuming(topic string) error {
	return conn.multiplexer.stopConsuming(topic, conn.name)
}

// SendSyncByte sends a message that uses byte encoder using the sync API
func (conn *Connection) SendSyncByte(topic string, key []byte, value []byte) (partition int32, offset int64, err error) {
	return conn.SendSyncMessage(topic, sarama.ByteEncoder(key), sarama.ByteEncoder(value))
}

// SendSyncString sends a message that uses string encoder using the sync API
func (conn *Connection) SendSyncString(topic string, key string, value string) (partition int32, offset int64, err error) {
	return conn.SendSyncMessage(topic, sarama.StringEncoder(key), sarama.StringEncoder(value))
}

//SendSyncMessage sends a message using the sync API
func (conn *Connection) SendSyncMessage(topic string, key client.Encoder, value client.Encoder) (partition int32, offset int64, err error) {
	msg, err := conn.multiplexer.syncProducer.SendMsg(topic, key, value)
	if err != nil {
		return 0, 0, err
	}
	return msg.Partition, msg.Offset, err
}

// SendAsyncByte sends a message that uses byte encoder using the async API
func (conn *Connection) SendAsyncByte(topic string, key []byte, value []byte, meta interface{}, successChan chan *client.ProducerMessage, errChan chan *client.ProducerError) {
	conn.SendAsyncMessage(topic, sarama.ByteEncoder(key), sarama.ByteEncoder(value), meta, successChan, errChan)
}

// SendAsyncString sends a message that uses string encoder using the async API
func (conn *Connection) SendAsyncString(topic string, key string, value string, meta interface{}, successChan chan *client.ProducerMessage, errChan chan *client.ProducerError) {
	conn.SendAsyncMessage(topic, sarama.StringEncoder(key), sarama.StringEncoder(value), meta, successChan, errChan)
}

// SendAsyncMessage sends a message using the async API
func (conn *Connection) SendAsyncMessage(topic string, key client.Encoder, value client.Encoder, meta interface{}, successChan chan *client.ProducerMessage, errChan chan *client.ProducerError) {
	auxMeta := &asyncMeta{successChan: successChan, errorChan: errChan, usersMeta: meta}
	conn.multiplexer.asyncProducer.SendMsg(topic, key, value, auxMeta)
}

// SendSyncMessage sends a message using the sync API
func (conn *ProtoConnection) SendSyncMessage(topic string, key string, value proto.Message) (partition int32, offset int64, err error) {
	data, err := conn.serializer.Marshal(value)
	if err != nil {
		return 0, 0, err
	}
	msg, err := conn.multiplexer.syncProducer.SendMsg(topic, sarama.StringEncoder(key), sarama.ByteEncoder(data))
	if err != nil {
		return 0, 0, err
	}
	return msg.Partition, msg.Offset, err
}

// SendAsyncMessage sends a message using the async API
func (conn *ProtoConnection) SendAsyncMessage(topic string, key string, value proto.Message, meta interface{}, successChan chan *client.ProducerMessage, errChan chan *client.ProducerError) error {
	data, err := conn.serializer.Marshal(value)
	if err != nil {
		return err
	}
	auxMeta := &asyncMeta{successChan: successChan, errorChan: errChan, usersMeta: meta}
	conn.multiplexer.asyncProducer.SendMsg(topic, sarama.StringEncoder(key), sarama.ByteEncoder(data), auxMeta)
	return nil
}

// ConsumeTopic is called to start consuming given topics.
// Function can be called until the multiplexer is started, it returns an error otherwise.
// The provided channel should be buffered, otherwise messages might be lost.
func (conn *ProtoConnection) ConsumeTopic(msgChan chan *client.ProtoConsumerMessage, topics ...string) error {
	conn.multiplexer.rwlock.Lock()
	defer conn.multiplexer.rwlock.Unlock()

	if conn.multiplexer.started {
		return fmt.Errorf("ConsumeTopic can be called only if the multiplexer has not been started yet")
	}

	internalChannel := make(chan *client.ConsumerMessage)

	go func() {
	messageHandler:
		for {
			select {
			case msg := <-internalChannel:
				select {
				case msgChan <- client.NewProtoConsumerMessage(msg, conn.serializer):
				default:
					log.Warn("Unable to deliver message to consumer")
				}
			case <-conn.multiplexer.consumer.GetCloseChannel():
				break messageHandler
			}
		}
		close(internalChannel)
	}()

	for _, topic := range topics {
		// check if we have already consumed the topic and partition
		subs, found := conn.multiplexer.mapping[topic]

		if !found {
			subs = &map[string]chan *client.ConsumerMessage{}
			conn.multiplexer.mapping[topic] = subs
		}
		// add subscription to consumerList
		(*subs)[conn.name] = internalChannel
		conn.multiplexer.mapping[topic] = subs
	}
	return nil
}

// StopConsuming cancels the previously created subscription for consuming the topic.
func (conn *ProtoConnection) StopConsuming(topic string) error {
	return conn.multiplexer.stopConsuming(topic, conn.name)
}
