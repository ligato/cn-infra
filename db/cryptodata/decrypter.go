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
	"encoding/json"
	"strings"
	"encoding/base64"
)

// DecryptFunc is function that decrypts arbitrary data
type DecryptFunc func(inData []byte) (data [] byte, err error)

// ValidateFunc is used to validate compatibility of JSON data when decrypting
type ValidateFunc func(inData []byte) (isValid bool)

// EncryptionCheck is used to check for data to contain encrypted marker
type EncryptionCheck struct {
	// IsEncrypted returns true if data was marked as encrypted
	IsEncrypted bool `json:"encrypted"`
}

// ArbitraryDecrypter is interface for decrypting groups of data
type ArbitraryDecrypter interface {
	// Decrypt decrypts input data using provided decrypting function for arbitrary data
	Decrypt(inData []byte, decryptFunc DecryptFunc) (data []byte, err error)
}

// DecrypterJSON is ArbitraryDecrypter implementations that can decrypt JSON values
type DecrypterJSON struct {
	// Prefix that is required for matching and decrypting values
	Prefix string
	// Validate validates data to be decrypted
	Validate ValidateFunc
}

// NewDecrypterJSON creates new JSON decrypter with default values for Prefix and Validate being
// `$crypto$` and presence of `encrypted: true`
func NewDecrypterJSON() *DecrypterJSON {
	return &DecrypterJSON{
		Prefix: "$crypto$",
		Validate: func(inData []byte) (isValid bool) {
			var jsonData EncryptionCheck
			err := json.Unmarshal(inData, &jsonData)
			return err == nil && jsonData.IsEncrypted
		},
	}
}

// Decrypt tries to decrypt JSON data that are encrypted based on `Validate` function return true on data.
// Then it parses data as JSON as tries to lookup all values that begin with `Prefix`, then trim prefix, base64
// decode the data and decrypt them using provided decrypt function.
func (d DecrypterJSON) Decrypt(inData []byte, decryptFunc DecryptFunc) ([]byte, error) {
	if !d.Validate(inData) {
		return inData, nil
	}

	var jsonData map[string]interface{}
	err := json.Unmarshal(inData, &jsonData)
	if err != nil {
		return nil, err
	}

	jsonData, err = d.decryptJSON(jsonData, decryptFunc)
	if err != nil {
		return nil, err
	}

	return json.Marshal(jsonData)
}

// decryptJSON recursively navigates JSON structure and tries to decrypt all string values with Prefix
func (d DecrypterJSON) decryptJSON(data map[string]interface{}, decryptFunc DecryptFunc) (map[string]interface{}, error) {
	for k, v := range data {
		switch t := v.(type) {
		case string:
			if s := strings.TrimPrefix(t, d.Prefix); s != t {
				s, err := base64.URLEncoding.DecodeString(s)
				if err != nil {
					return nil, err
				}

				arbitraryData, err := decryptFunc(s)
				if err != nil {
					return nil, err
				}

				data[k] = string(arbitraryData)
			}
		case map[string]interface{}:
			v, err := d.decryptJSON(t, decryptFunc)
			if err != nil {
				return nil, err
			}

			data[k] = v
		}
	}

	return data, nil
}
