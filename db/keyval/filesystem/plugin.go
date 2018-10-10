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

package filesystem

import (
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/filesystem/reader"
	"github.com/ligato/cn-infra/db/keyval/kvproto"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/servicelabel"
)

// Plugin filesystem uses host os file system as database to store configuration.
type Plugin struct {
	Deps

	// Filesystem client
	client *Client

	// Plugin config
	config *Config
	// Plugin is disabled without config
	disabled bool
	// Read or write proto modelled data
	protoWrapper *kvproto.ProtoWrapper
}

type Deps struct {
	infra.PluginDeps
	sv servicelabel.ReaderAPI
}

type Config struct {
	Paths []string `json:"paths"`
}

// Init reads file config and creates new client to communicate with file system
func (p *Plugin) Init() error {
	// Read filesystem configuration file
	var err error
	p.config, err = p.getFilesystemConfig()
	if err != nil || p.disabled {
		return err
	}

	if p.client, err = NewClient(p.config.Paths, p.sv.GetAgentPrefix(), reader.NewReader(), p.Log); err != nil {
		return err
	}

	p.protoWrapper = kvproto.NewProtoWrapper(p.client, &keyval.SerializerJSON{})

	return nil
}

// AfterInit starts file system event watcher
func (p *Plugin) AfterInit() error {
	go p.client.eventWatcher()

	return nil
}

func (p *Plugin) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

func (p *Plugin) Disabled() bool {
	return p.disabled
}

func (p *Plugin) OnConnect(callback func() error) {
	if err := callback(); err != nil {
		p.Log.Error(err)
	}
}

func (p *Plugin) String() string {
	return p.PluginName.String()
}

func (p *Plugin) NewBroker(keyPrefix string) keyval.ProtoBroker {
	return p.protoWrapper.NewBroker(keyPrefix)
}

func (p *Plugin) NewWatcher(keyPrefix string) keyval.ProtoWatcher {
	return p.protoWrapper.NewWatcher(keyPrefix)
}

func (p *Plugin) getFilesystemConfig() (*Config, error) {
	var filesystemCfg Config
	found, err := p.Cfg.LoadValue(&filesystemCfg)
	if err != nil {
		return nil, err
	}
	if !found {
		p.Log.Warnf("Filesystem config not found, skip loading this plugin")
		p.disabled = true
	}
	return &filesystemCfg, nil
}
