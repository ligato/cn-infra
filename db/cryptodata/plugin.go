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
	"github.com/ligato/cn-infra/flavors/local"
	"crypto/rsa"
	"github.com/ligato/cn-infra/db/keyval"
	"crypto/rand"
	"github.com/pkg/errors"
)

// Deps lists dependencies of the cryptodata plugin.
type Deps struct {
	local.PluginInfraDeps
}

// Plugin implements cryptodata as plugin.
type Plugin struct {
	Deps
	// Plugin is disabled if there is no config file available
	disabled bool
	// List of private keys required from config
	privateKeys []*rsa.PrivateKey
}

// Init initializes cryptodata plugin.
func (plugin *Plugin) Init() (err error) {
	var config Config
	found, err := plugin.PluginConfig.GetValue(&config)
	if err != nil {
		return err
	}

	if !found {
		plugin.Log.Info("cryptodata config not found, skip loading this plugin")
		plugin.disabled = true
		return nil
	}

	return plugin.UpdateConfig(config)
}

// Disabled returns *true* if the plugin is not in use due to missing configuration.
func (plugin *Plugin) Disabled() bool {
	return plugin.disabled
}

// UpdateConfig updates private key configuration for plugin
func (plugin *Plugin) UpdateConfig(config Config) (err error) {
	var clientConfig ClientConfig
	err = ReadCryptoConfig(config, &clientConfig)
	if err == nil {
		plugin.privateKeys = clientConfig.PrivateKeys
	}
	return
}

// EncryptArbitrary decrypts arbitrary input data
func (plugin *Plugin) EncryptArbitrary(inData []byte) (data []byte, err error) {
	for _, key := range plugin.privateKeys {
		data, err := rsa.EncryptPKCS1v15(rand.Reader, &key.PublicKey, inData)

		if err == nil {
			return data, nil
		}
	}

	return nil, errors.New("Failed to encrypt data due to all private keys failing")
}

// DecryptArbitrary decrypts arbitrary input data
func (plugin *Plugin) DecryptArbitrary(inData []byte) (data []byte, err error) {
	for _, key := range plugin.privateKeys {
		data, err := rsa.DecryptPKCS1v15(rand.Reader, key, data)

		if err == nil {
			return data, nil
		}
	}

	return nil, errors.New("Failed to decrypt data due to no private key matching")
}

// Wrap wraps core broker watcher with support for decrypting encrypted keys
func (plugin *Plugin) Wrap(cbw keyval.CoreBrokerWatcher, decrypter Decrypter) keyval.CoreBrokerWatcher {
	return keyval.CoreBrokerWatcher(NewCoreBrokerWatcherWrapper(cbw, decrypter, plugin.DecryptArbitrary))
}
