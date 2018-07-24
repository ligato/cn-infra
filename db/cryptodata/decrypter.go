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

const cryptoPrefix = "$crypto$"
var encryptedJSONRegex = regexp.MustCompile(`^{\s*"encrypted"\s*:\s*"true"`)


// DecryptFunc is function that decrypts arbitrary data
type DecryptFunc func(inData []byte) (data [] byte, err error)

// ArbitraryDecrypter is interface for decrypting groups of data
type ArbitraryDecrypter interface {
	// Decrypt decrypts input data using provided decrypting function for arbitrary data
	Decrypt(inData []byte, decryptFunc DecryptFunc) (data []byte, err error)
}

// DecrypterJSON is ArbitraryDecrypter implementations that can decrypt JSON values
type DecrypterJSON struct {
	// Prefix that is required for matching and decrypting values
	Prefix string
	// Present checks if JSON string matches regex in order to start decrypting (and exit early if not)
	Present *regexp.Regexp
}

// NewDecrypterJSON creates new JSON decrypter with default values for Prefix and Present being
// $crypto$ and ^{\s*"encrypted"\s*:\s*"true"
func NewDecrypterJSON() *DecrypterJSON {
	return &DecrypterJSON{
		Prefix: cryptoPrefix,
		Present: encryptedJSONRegex,
	}
}

// Decrypt tries to decrypt JSON data that are encrypted based on "encrypted": "true" presence in it.
// Then it parses data as JSON as tries to lookup all top-level values that begin with $crypto$ and decrypt them
// using provided arbitrary decrypt function.
func (d DecrypterJSON) Decrypt(inData []byte, decryptFunc DecryptFunc) (data []byte, err error) {
	data = inData

	if !d.Present.Match(inData) {
		return
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(inData, &jsonData)
	if err != nil {
		return
	}

	jsonData = d.decryptJSON(jsonData, decryptFunc)
	data, err = json.Marshal(jsonData)
	return
}

// decryptJSON recursively navigates JSON structure and tries to decrypt all string values with Prefix
func (d DecrypterJSON) decryptJSON(data map[string]interface{}, decryptFunc DecryptFunc) map[string]interface{} {
	for k, v := range data {
		var stringVal string
		switch t := v.(type) {
		case string:
			stringVal = t
		case map[string]interface{}:
			v = d.decryptJSON(t, decryptFunc)
			continue
		default:
			continue
		}

		if !strings.HasPrefix(stringVal, d.Prefix) {
			continue
		}

		stringVal = strings.TrimPrefix(stringVal, d.Prefix)
		arbitraryData, err := decryptFunc([]byte(stringVal))
		if err == nil {
			data[k] = string(arbitraryData)
		}
	}

	return data
}