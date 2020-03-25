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

// +build !redact

// Package redact provides utilities for redacting sensitive data.
package redact

import "os"

// Enabled can be used to disable redacting.
var enabled = true

func init() {
	if os.Getenv("NOREDACT") != "" {
		enabled = false
	}
}

func SetEnabled(b bool) {
	enabled = b
}
