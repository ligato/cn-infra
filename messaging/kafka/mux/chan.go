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
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/messaging"
	"github.com/ligato/cn-infra/messaging/kafka/client"
	"time"
)

// DefaultMsgTimeout for delivery of notification
const DefaultMsgTimeout = 2 * time.Second

// ToProtoMsgChan allows to receive messages through channel instead of callback.
func ToProtoMsgChan(ch chan messaging.ProtoMessage, opts ...interface{}) func(messaging.ProtoMessage) {

	timeout, logger := parseOpts(opts...)

	return func(msg messaging.ProtoMessage) {
		select {
		case ch <- msg:
		case <-time.After(timeout):
			logger.Warn("Unable to deliver message")
		}
	}
}

// ToProtoMsgErrChan allows to receive error messages through channel instead of callback.
func ToProtoMsgErrChan(ch chan messaging.ProtoMessageErr, opts ...interface{}) func(messaging.ProtoMessageErr) {

	timeout, logger := parseOpts(opts...)

	return func(msg messaging.ProtoMessageErr) {
		select {
		case ch <- msg:
		case <-time.After(timeout):
			logger.Warn("Unable to deliver message")
		}
	}
}

// ToBytesMsgChan allows to receive ConsumerMessage through channel. This function can be used as an argument for
// ConsumeTopic call.
func ToBytesMsgChan(ch chan *client.ConsumerMessage, opts ...interface{}) func(*client.ConsumerMessage) {

	timeout, logger := parseOpts(opts...)

	return func(msg *client.ConsumerMessage) {
		select {
		case ch <- msg:
		case <-time.After(timeout):
			logger.Warn("Unable to deliver message")
		}
	}
}

// ToBytesProducerChan allows to receive ProducerMessage through channel. This function can be used as an argument for
// methods publishing using async API.
func ToBytesProducerChan(ch chan *client.ProducerMessage, opts ...interface{}) func(*client.ProducerMessage) {

	timeout, logger := parseOpts(opts...)

	return func(msg *client.ProducerMessage) {
		select {
		case ch <- msg:
		case <-time.After(timeout):
			logger.Warn("Unable to deliver message")
		}
	}
}

// ToBytesProducerErrChan allows to receive ProducerMessage through channel. This function can be used as an argument for
// methods publishing using async API.
func ToBytesProducerErrChan(ch chan *client.ProducerError, opts ...interface{}) func(*client.ProducerError) {

	timeout, logger := parseOpts(opts...)

	return func(msg *client.ProducerError) {
		select {
		case ch <- msg:
		case <-time.After(timeout):
			logger.Warn("Unable to deliver message")
		}
	}
}

func parseOpts(opts ...interface{}) (time.Duration, logging.Logger) {
	timeout := DefaultMsgTimeout
	logger := logroot.Logger()

	for _, opt := range opts {
		switch opt.(type) {
		case *WithLoggerOpt:
			logger = opt.(*WithLoggerOpt).logger
		case *WithTimeoutOpt:
			timeout = opt.(*WithTimeoutOpt).timeout
		}
	}
	return timeout, logger

}

// WithTimeoutOpt defines the maximum time that is attempted to deliver notification.
type WithTimeoutOpt struct {
	timeout time.Duration
}

// WithTimeout creates an option for ToChan function that defines a timeout for notification delivery.
func WithTimeout(timeout time.Duration) *WithTimeoutOpt {
	return &WithTimeoutOpt{timeout: timeout}
}

// WithLoggerOpt defines a logger that logs if delivery of notification is unsuccessful.
type WithLoggerOpt struct {
	logger logging.Logger
}

// WithLogger creates an option for ToChan function that specifies a logger to be used.
func WithLogger(logger logging.Logger) *WithLoggerOpt {
	return &WithLoggerOpt{logger: logger}
}
