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
type DecryptFunc func(inData []byte) (data []byte, err error)

// CheckEncryptedFunc is used to check if JSON data is encrypted
type CheckEncryptedFunc func(inData []byte) (isEncrypted bool)

// EncryptionCheck is used to check for data to contain encrypted marker
type EncryptionCheck struct {
	// IsEncrypted returns true if data was marked as encrypted
	IsEncrypted bool `json:"encrypted"`
}

// ArbitraryDecrypter represents decrypter that looks for encrypted values inside arbitrary data and returns
// the data with the values decrypted
type ArbitraryDecrypter interface {
	// Decrypt decrypts input data using provided decrypting function for arbitrary data
	Decrypt(inData []byte, decryptFunc DecryptFunc) (data []byte, err error)
}

// DecrypterJSON is ArbitraryDecrypter implementations that can decrypt JSON values
type DecrypterJSON struct {
	// Prefix that is required for matching and decrypting values
	Prefix string
	// CheckEncrypted checks if provided data are marked as encrypted.
	// By default this function tries to unmarshal JSON to EncryptionCheck and check the IsEncrypted being true
	CheckEncrypted CheckEncryptedFunc
}

// NewDecrypterJSON creates new JSON decrypter with default values for Prefix and CheckEncrypted being
// `$crypto$` and presence of `encrypted: true`
func NewDecrypterJSON() *DecrypterJSON {
	return &DecrypterJSON{
		Prefix: "$crypto$",
		CheckEncrypted: func(inData []byte) (isValid bool) {
			var jsonData EncryptionCheck
			err := json.Unmarshal(inData, &jsonData)
			return err == nil && jsonData.IsEncrypted
		},
	}
}


// Decrypt tries to find encrypted values in JSON data and decrypt them. It uses CheckEncrypted function on the
// data to check if it contains any encrypted data.
// Then it parses data as JSON as tries to lookup all values that begin with `Prefix`, then trim prefix, base64
// decode the data and decrypt them using provided decrypt function.
func (d DecrypterJSON) Decrypt(inData []byte, decryptFunc DecryptFunc) ([]byte, error) {
	if !d.CheckEncrypted(inData) {
		return inData, nil
	}

	var jsonData map[string]interface{}
	err := json.Unmarshal(inData, &jsonData)
	if err != nil {
		return nil, err
	}

	err = d.decryptJSON(jsonData, decryptFunc)
	if err != nil {
		return nil, err
	}

	return json.Marshal(jsonData)
}

// decryptJSON recursively navigates JSON structure and tries to decrypt all string values with Prefix
func (d DecrypterJSON) decryptJSON(data map[string]interface{}, decryptFunc DecryptFunc) (error) {
	for k, v := range data {
		switch t := v.(type) {
		case string:
			if s := strings.TrimPrefix(t, d.Prefix); s != t {
				s, err := base64.URLEncoding.DecodeString(s)
				if err != nil {
					return err
				}

				decryptedData, err := decryptFunc(s)
				if err != nil {
					return err
				}

				data[k] = string(decryptedData)
			}
		case map[string]interface{}:
			err := d.decryptJSON(t, decryptFunc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
