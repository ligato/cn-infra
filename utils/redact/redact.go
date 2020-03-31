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

// Package redact provides utilities for redacting sensitive data.
package redact

import (
	"github.com/golang/protobuf/proto"
)

// StringRedactor is the default redactor for strings.
var StringRedactor = func(s string) string {
	return MaskedString(s).String()
}

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
		return redactProto(x)
	}

	return v
}

// String returns redacted string.
func String(s string) string {
	if !enabled {
		return s
	}

	return StringRedactor(s)
}
