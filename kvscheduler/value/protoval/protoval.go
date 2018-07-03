package protoval

import (
	"github.com/gogo/protobuf/proto"
	. "github.com/ligato/cn-infra/kvscheduler/api"
)

// ProtoValue is an interface that value carrying proto message should implement.
type ProtoValue interface {
	Value
	GetProtoMessage() proto.Message
}

// protoValue wraps ProtoMessage to implement the Value interface for use with
// the KVScheduler.
type protoValue struct {
	protoMessage proto.Message
}

// ProtoMessageWithName is based on our convention for proto-defined objects
// to store object label under the attribute "Name".
type ProtoMessageWithName interface {
	GetName() string
}

// NewProtoValue creates a new instance of ProtoValue carrying the given proto
// message.
func NewProtoValue(protoMsg proto.Message) ProtoValue {
	if protoMsg == nil {
		return nil
	}
	return &protoValue{protoMessage: protoMsg}
}

// GetProtoMessage returns the underlying proto message.
func (pv * protoValue) GetProtoMessage() proto.Message {
	return pv.protoMessage
}

// Label tries to read and return "Name" attribute. Without the name attribute,
// the function will return empty string.
func (pv * protoValue) Label() string {
	protoWithName, hasName := pv.protoMessage.(ProtoMessageWithName)
	if hasName {
		return protoWithName.GetName()
	}
	return pv.String()
}

// Equivalent uses proto.Equal for comparison.
func (pv * protoValue) Equivalent(v2 Value) bool {
	v2Proto, ok := v2.(*protoValue)
	if !ok {
		return false
	}
	return proto.Equal(pv.protoMessage, v2Proto.protoMessage)
}

// String uses the String method from proto.Message.
func (pv * protoValue) String() string {
	return pv.protoMessage.String()
}

// Type returns Object.
func (pv * protoValue) Type() ValueType {
	return Object
}