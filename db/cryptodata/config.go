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
	"crypto/rsa"
	"io/ioutil"
	"encoding/pem"
	"crypto/x509"
	"errors"
)

// Config is used to read private key from file
type Config struct {
	// Private key file is used to create rsa.PrivateKey from this PEM path
	PrivateKeyFiles []string `json:"private-key-files"`
}

// ClientConfig is result of converting Config.PrivateKeyFile to PrivateKey
type ClientConfig struct {
	// Private key is used to decrypt encrypted keys while reading them from store
	PrivateKeys []*rsa.PrivateKey
}

// ReadCryptoConfig reads private key from PEM path
func ReadCryptoConfig(config Config) (clientConfig ClientConfig, err error) {
	clientConfig = ClientConfig{}

	if len(config.PrivateKeyFiles) == 0 {
		return
	}

	for _, file := range config.PrivateKeyFiles {
		bytes, err := ioutil.ReadFile(file)
		if err != nil {
			return clientConfig, err
		}

		block, _ := pem.Decode(bytes)
		if block == nil {
			return clientConfig, errors.New("failed to decode PEM for key " + file)
		}

		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return clientConfig, err
		}

		privateKey.Precompute()
		err = privateKey.Validate()
		if err != nil {
			return clientConfig, err
		}

		clientConfig.PrivateKeys = append(clientConfig.PrivateKeys, privateKey)
	}

	return
}
