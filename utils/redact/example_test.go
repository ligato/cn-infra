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

	"go.ligato.io/cn-infra/v2/utils/redact"
)

type Data struct {
	Username string
	Password string
}

func (d Data) Redacted() interface{} {
	d.Password = redact.String(d.Password)
	return d
}

func Example_value() {
	data := Data{
		Username: "bob",
		Password: "password123",
	}
	fmt.Printf("%+v", redact.Value(data))
	// Output: {Username:bob Password:***********}
}
