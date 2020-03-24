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
	"testing"
)

func TestMaskString(t *testing.T) {
	var input = "data"
	const redactedStr = "****"

	clear := fmt.Sprintf("%s", input)
	if clear != input {
		t.Fatalf("expected:\n%q, got\n%q", input, clear)
	}

	secret := fmt.Sprintf("%s", BlackedString(input))
	if secret != redactedStr {
		t.Fatalf("expected:\n%q, got\n%q", redactedStr, secret)
	}

	type Data struct {
		Clear  string
		Secret BlackedString
	}
	data := Data{"clear", "secret"}
	const redactedData = `{Clear:clear Secret:******}`

	out := fmt.Sprintf("%+v", data)
	if out != redactedData {
		t.Fatalf("expected:\n%q, got\n%q", redactedData, out)
	}
}
