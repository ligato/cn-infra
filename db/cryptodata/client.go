//  Copyright (c) 2018 Cisco and/or its affiliates.
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

package cryptodata

import (
	"crypto/rsa"
	"github.com/ligato/cn-infra/db/keyval"
	"errors"
	"crypto/rand"
	"io"
)

// ClientConfig is result of converting Config.PrivateKeyFile to PrivateKey
type ClientConfig struct {
	// Private key is used to decrypt encrypted keys while reading them from store
	PrivateKeys []*rsa.PrivateKey
	// Reader used for encrypting/decrypting
	Reader io.Reader
}

// Client handles encrypting/decrypting and wrapping data
type Client struct {
	ClientConfig
}

// NewClient creates new client from provided config and reader
func NewClient(clientConfig ClientConfig) (client *Client) {
	client = &Client{
		ClientConfig: clientConfig,
	}

	// If reader is nil use default rand.Reader
	if client.Reader == nil {
		client.Reader = rand.Reader
	}

	return
}

// EncryptArbitrary decrypts arbitrary input data using provided public key
func (client *Client) EncryptArbitrary(inData []byte, pub *rsa.PublicKey) (data []byte, err error) {
	data, err = rsa.EncryptPKCS1v15(client.Reader, pub, inData)
	return
}

// DecryptArbitrary decrypts arbitrary input data
func (client *Client) DecryptArbitrary(inData []byte) (data []byte, err error) {
	for _, key := range client.PrivateKeys {
		data, err := rsa.DecryptPKCS1v15(client.Reader, key, inData)

		if err == nil {
			return data, nil
		}
	}

	return nil, errors.New("failed to decrypt data due to no private key matching")
}

// Wrap wraps core broker watcher with support for decrypting encrypted keys
func (client *Client) Wrap(cbw keyval.CoreBrokerWatcher, decrypter Decrypter) keyval.CoreBrokerWatcher {
	return keyval.CoreBrokerWatcher(NewCoreBrokerWatcherWrapper(cbw, decrypter, client.DecryptArbitrary))
}
