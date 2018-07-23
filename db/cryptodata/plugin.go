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
)

// Deps lists dependencies of the cryptodata plugin.
type Deps struct {
	local.PluginInfraDeps
}

// Plugin implements cryptodata as plugin.
type Plugin struct {
	Deps
	// Client provides crypto support
	Client
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

	plugin.Client, err = NewClient(config)
	return
}

// Disabled returns *true* if the plugin is not in use due to missing configuration.
func (plugin *Plugin) Disabled() bool {
	return plugin.disabled
}
