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
	"reflect"
	"testing"
)

func TestRedactMapping(t *testing.T) {
	type Nested struct {
		Hidden string
	}
	type MyType struct {
		Name   string
		Secret string
		Nest   Nested
	}

	tests := []struct {
		name   string
		init   func()
		input  MyType
		expect MyType
	}{
		{
			"trivial",
			func() { RegisterFieldsPaths(MyType{}, []string{"Secret"}) },
			MyType{Name: "bla", Secret: "secret"},
			MyType{Name: "bla", Secret: "******"},
		},
		{
			"nested",
			func() { RegisterFieldsPaths(MyType{}, []string{"Nest", "Hidden"}) },
			MyType{Name: "bla", Nest: Nested{"hidden"}},
			MyType{Name: "bla", Nest: Nested{"******"}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.init != nil {
				test.init()
			}
			data := test.input

			//t.Logf("before: %T %+v", data, data)
			redactMapped(&data)
			//t.Logf(" after: %T %+v", data, data)

			if !reflect.DeepEqual(data, test.expect) {
				t.Fatalf("expected:\n%#v, got:\n%#v", test.expect, data)
			}
		})
	}
}
