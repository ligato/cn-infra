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

var redactorType = reflect.TypeOf((*Redactor)(nil)).Elem()
var protoMsgType = reflect.TypeOf((*proto.Message)(nil)).Elem()

func redactProto(msg proto.Message) proto.Message {
	if !ContainsRedacted(msg) {
		return msg
	}

	msgCopy := proto.Clone(msg)
	val := reflect.ValueOf(msgCopy)
	redactFields(val)
	return msgCopy
}

func ContainsRedacted(x interface{}) bool {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("must be Ptr kind, not %v", v.Kind()))
	}
	r := v
	return containsRedactedField(r)
}

func containsRedactedField(r reflect.Value) bool {
	check := func(v reflect.Value) bool {
		if v.Type().Implements(redactorType) {
			return true
		}
		if v.Type().Implements(protoMsgType) {
			if b := containsRedactedField(v); b {
				return true
			}
		}
		return false
	}
	for i := 0; i < r.Elem().NumField(); i++ {
		f := r.Elem().Field(i)
		switch f.Kind() {
		case reflect.Slice:
			for i := 0; i < f.Len(); i++ {
				sv := f.Index(i)
				if contains := check(sv); contains {
					return true
				}
			}
		case reflect.Ptr:
			if f.IsNil() {
				continue
			}
		default:
			continue
		}
		if contains := check(f); contains {
			return true
		}
	}
	return false
}

func redactFields(r reflect.Value) {
	check := func(v reflect.Value) {
		if v.Type().Implements(redactorType) {
			rv := v.Interface().(Redactor).Redacted()
			v.Elem().Set(reflect.ValueOf(rv).Elem())
			return
		}
		if v.Type().Implements(protoMsgType) {
			redactFields(v)
		}
	}
	for i := 0; i < r.Elem().NumField(); i++ {
		f := r.Elem().Field(i)
		switch f.Kind() {
		case reflect.Slice:
			for i := 0; i < f.Len(); i++ {
				sv := f.Index(i)
				check(sv)
			}
		case reflect.Ptr:
			if f.IsNil() {
				continue
			}
		default:
			continue
		}
		check(f)
		/*if f.Kind() != reflect.Ptr || f.IsNil() {
			continue
		}
		if f.Type().Implements(redactorType) {
			rv := f.Interface().(Redactor).Redacted()
			f.Elem().Set(reflect.ValueOf(rv).Elem())
			continue
		}
		if f.Type().Implements(protoMsgType) {
			redactFields(f)
		}*/
	}
}
