//  Copyright (c) 2020 Cisco and/or its affiliates.
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

package redact

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
)

// Redactor is an interface to be implemented by types that contain
// sensitive data that should be redacted. Redacted should return copy
// of the value with the sensitive fields redacted.
type Redactor interface {
	Redacted() interface{}
}

// Value returns value with sensitive fields redacted.
// The value will return unchanged if it cannot be redacted.
func Value(v interface{}) interface{} {
	if !enabled {
		return v
	}

	switch x := v.(type) {
	case Redactor:
		return x.Redacted()
	case proto.Message:
		return protoValue(x)
	}
	if isMapped(v) {
		redactMapped(v)
	}

	return v
}

func protoValue(msg proto.Message) proto.Message {
	if !containsRedacted(msg) {
		return msg
	}
	msgCopy := proto.Clone(msg)
	val := reflect.ValueOf(msgCopy)
	redactFields(val)
	return msgCopy
}

func containsRedacted(x interface{}) bool {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("must be Ptr kind, not %v", v.Kind()))
	}
	r := v
	return containsRedactedField(r)
}

var redactorType = reflect.TypeOf((*Redactor)(nil)).Elem()
var protoMsgType = reflect.TypeOf((*proto.Message)(nil)).Elem()

func containsRedactedField(r reflect.Value) bool {
	for i := 0; i < r.Elem().NumField(); i++ {
		f := r.Elem().Field(i)
		if f.Kind() != reflect.Ptr || f.IsNil() {
			continue
		}
		if f.Type().Implements(redactorType) {
			return true
		}
		if f.Type().Implements(protoMsgType) {
			if b := containsRedactedField(f); b {
				return true
			}
		}
	}
	return false
}

func redactFields(r reflect.Value) {
	for i := 0; i < r.Elem().NumField(); i++ {
		f := r.Elem().Field(i)
		if f.Kind() != reflect.Ptr || f.IsNil() {
			continue
		}
		if f.Type().Implements(redactorType) {
			rv := f.Interface().(Redactor).Redacted()
			f.Elem().Set(reflect.ValueOf(rv).Elem())
			continue
		}
		if f.Type().Implements(protoMsgType) {
			redactFields(f)
		}
	}
}
