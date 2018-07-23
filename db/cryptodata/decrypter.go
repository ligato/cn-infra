// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cryptodata

import (
	"regexp"
	"encoding/json"
	"strings"
)

var encryptedJSONRegex = regexp.MustCompile(`"encrypted"\s*:\s*"true"`)

// DecryptArbitrary is function that decrypts arbitrary data
type DecryptArbitrary func(inData []byte) (data [] byte, err error)

// Decrypter is interface for decrypting groups of data
type Decrypter interface {
	// Decrypt decrypts input data using provided decrypting function for arbitrary data
	Decrypt(inData []byte, decryptArbitrary DecryptArbitrary) (data []byte)
}

// DecrypterJSON is Decrypter implementations that can decrypt JSON values
type DecrypterJSON struct{}

// Decrypt tries to decrypt JSON data that are encrypted based on "encrypted": "true" presence in it.
// Then it parses data as JSON as tries to lookup all top-level values that begin with $crypto$ and decrypt them
// using provided arbitrary decrypt function.
func (DecrypterJSON) Decrypt(inData []byte, decryptArbitrary DecryptArbitrary) (data []byte) {
	data = inData

	if !encryptedJSONRegex.Match(inData) {
		return
	}

	var jsonData interface{}
	err := json.Unmarshal(inData, &jsonData)
	if err != nil {
		return
	}

	jsonMap, ok := jsonData.(map[string]interface{})
	if !ok {
		return
	}

	for k, v := range jsonMap {
		stringVal, ok := v.(string)
		prefix := "$crypto$"

		if !ok || !strings.HasPrefix(stringVal, prefix) {
			continue
		}

		stringVal = strings.TrimPrefix(stringVal, prefix)
		arbitraryData, err := decryptArbitrary([]byte(stringVal))
		if err == nil {
			jsonMap[k] = string(arbitraryData)
		}
	}

	jsonBytes, err := json.Marshal(jsonMap)
	if err == nil {
		data = jsonBytes
	}

	return
}
