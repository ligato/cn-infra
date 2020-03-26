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
	"fmt"
	"testing"

	"go.ligato.io/cn-infra/v2/utils/redact"
	"go.ligato.io/cn-infra/v2/utils/redact/testdata"
)

func TestValue(t *testing.T) {
	data := Data{
		Username: "bob",
		Password: "password123",
	}
	out := fmt.Sprint(redact.Value(data))
	const expected = `{bob ***********}`
	if out != expected {
		t.Fatalf("expected:\n%q, but got:\n%q", expected, out)
	}
}

func TestProto(t *testing.T) {
	data := &testdata.TestData{
		Username: "bob",
		Password: "password123",
	}
	out := fmt.Sprint(redact.Value(data))
	const expected = `username:"bob" password:"***********" `
	if out != expected {
		t.Fatalf("expected:\n%q, but got:\n%q", expected, out)
	}
}

func TestProtoNested(t *testing.T) {
	data := &testdata.TestNested{
		Name: "SomeName",
		Data: &testdata.TestData{
			Username: "bob",
			Password: "password123",
		},
	}
	// check if nested data is redacted
	out := fmt.Sprint(redact.Value(data))
	const expected = `name:"SomeName" data:<username:"bob" password:"***********" > `
	if out != expected {
		t.Fatalf("expected redacted fields:\n%q, but got:\n%q", expected, out)
	}
	// check if original data is unchanged
	out2 := fmt.Sprint(data)
	const expected2 = `name:"SomeName" data:<username:"bob" password:"password123" > `
	if out2 != expected2 {
		t.Fatalf("expected original data:\n%q, to not change but got:\n%q", expected2, out2)
	}
}

func TestContainsRedacted(t *testing.T) {
	{
		data := &testdata.TestData{
			Username: "bob",
			Password: "password123",
		}
		b := redact.ContainsRedacted(data)
		if b != false {
			t.Fatal("expected to not contain redacted")
		}
	}
	{
		data := &testdata.TestNested{
			Name: "SomeName",
			Data: &testdata.TestData{
				Username: "bob",
				Password: "password123",
			},
		}
		b := redact.ContainsRedacted(data)
		if b != true {
			t.Fatal("expected to contain redacted")
		}
	}
	{
		data := &testdata.TestNested{
			Name: "SomeName",
			Data: nil,
		}
		b := redact.ContainsRedacted(data)
		if b != false {
			t.Fatal("expected to not contain redacted")
		}
	}
}
