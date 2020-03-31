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

import "strings"

// MaskedString is a type that masks an actual string data.
// It implements fmt.Stringer and encoding.TextMarshaler for returning
// asterisks `*` characters with same length as the actual string value.
type MaskedString string

func (s MaskedString) Redacted() interface{} {
	return s.String()
}

func (s MaskedString) String() string {
	return strings.Repeat("*", len(s))
}

func (s MaskedString) MarshalText() (text []byte, err error) {
	return []byte(s.String()), nil
}
