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

package redact_test

import (
	"testing"

	. "go.ligato.io/cn-infra/v2/utils/redact"
	"go.ligato.io/cn-infra/v2/utils/redact/testdata"
)

func TestContainsRedacted(t *testing.T) {
	t.Run("straight", func(t *testing.T) {
		data := &testdata.TestData{
			Username: "bob",
			Password: "password123",
		}
		b := ContainsRedacted(data)
		if b != false {
			t.Fatal("expected to not contain redacted")
		}
	})
	t.Run("nested", func(t *testing.T) {
		data := &testdata.TestNested{
			Name: "SomeName",
			Data: &testdata.TestData{
				Username: "bob",
				Password: "password123",
			},
		}
		b := ContainsRedacted(data)
		if b != true {
			t.Fatal("expected to contain redacted")
		}
	})
	t.Run("nested nil", func(t *testing.T) {
		data := &testdata.TestNested{
			Name: "SomeName",
			Data: nil,
		}
		b := ContainsRedacted(data)
		if b != false {
			t.Fatal("expected to not contain redacted")
		}
	})
	t.Run("slice", func(t *testing.T) {
		data := &testdata.TestSlice{
			Name: "SomeName",
			Data: []*testdata.TestData{
				{
					Username: "bob",
					Password: "password123",
				},
				{
					Username: "alice",
					Password: "123456",
				},
			},
		}
		b := ContainsRedacted(data)
		if b != true {
			t.Fatal("expected to contain redacted")
		}
	})
}
