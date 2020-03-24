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
	"strings"
)

// StringRedactor is default redacting of sensitive strings.
var StringRedactor = func(s string) string {
	return strings.Repeat("*", len(s))
}

// Redactor is an interface to be implemented by types that contain
// sensitive data that should be redacted.
type Redactor interface {
	Redacted() interface{}
}

// Value returns value with sensitive fields redacted.
// The value will return unchanged if it cannot be redacted.
func Value(v interface{}) interface{} {
	if !enabled {
		return v
	}

	if r, ok := v.(Redactor); ok {
		v = r.Redacted()
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

// BlackedString blacks out strings.
type BlackedString string

func (s BlackedString) String() string {
	return strings.Repeat("*", len(s))
}
