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
	"strings"
	"testing"

	"go.ligato.io/cn-infra/v2/utils/redact"
	"go.ligato.io/cn-infra/v2/utils/redact/testdata"
)

type Data struct {
	Username string
	Password string
}

func (d Data) Redacted() interface{} {
	d.Password = redact.String(d.Password)
	return d
}

func TestString(t *testing.T) {
	const (
		pwd      = "password123"
		expected = "***********"
	)
	redacted := redact.String(pwd)
	if redacted != expected {
		t.Fatalf("expected redacted string:\n%q, but got:\n%q", expected, redacted)
	}
}

func TestValue(t *testing.T) {
	data := Data{
		Username: "bob",
		Password: "password123",
	}
	const (
		expected = `{bob ***********}`
	)
	out := fmt.Sprint(redact.Value(data))
	if out != expected {
		t.Fatalf("expected:\n%q, but got:\n%q", expected, out)
	}
}

func TestProto(t *testing.T) {
	data := &testdata.TestData{
		Username: "bob",
		Password: "password123",
	}
	const (
		expectedUser     = `username:"bob"`
		expectedPassword = `password:"***********"`
	)
	out := fmt.Sprint(redact.Value(data))
	if !strings.Contains(out, expectedUser) {
		t.Fatalf("expected to contain:\n%q, but got:\n%q", expectedUser, out)
	}
	if !strings.Contains(out, expectedPassword) {
		t.Fatalf("expected to contain:\n%q, but got:\n%q", expectedPassword, out)
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
	expectUnchanged := fmt.Sprint(data)
	expectRedacted := fmt.Sprint(&testdata.TestNested{
		Name: "SomeName",
		Data: &testdata.TestData{
			Username: "bob",
			Password: "***********",
		},
	})

	// check if nested data is redacted
	out := fmt.Sprint(redact.Value(data))
	if out != expectRedacted {
		t.Errorf("expected redacted fields:\n%q, but got:\n%q", expectRedacted, out)
	}
	// check if original data is unchanged
	out2 := fmt.Sprint(data)
	if out2 != expectUnchanged {
		t.Errorf("expected original data:\n%q, to not change but got:\n%q", expectUnchanged, out2)
	}
}

func TestProtoSlice(t *testing.T) {
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
	expectUnchanged := fmt.Sprint(data)
	expectRedacted := fmt.Sprint(&testdata.TestSlice{
		Name: "SomeName",
		Data: []*testdata.TestData{
			{
				Username: "bob",
				Password: "***********",
			},
			{
				Username: "alice",
				Password: "******",
			},
		},
	})

	// check if nested data is redacted
	out := fmt.Sprint(redact.Value(data))
	if out != expectRedacted {
		t.Errorf("expected redacted fields:\n%q, but got:\n%q", expectRedacted, out)
	}
	// check if original data is unchanged
	out2 := fmt.Sprint(data)
	if out2 != expectUnchanged {
		t.Errorf("expected original data:\n%q, to not change but got:\n%q", expectUnchanged, out2)
	}
}
