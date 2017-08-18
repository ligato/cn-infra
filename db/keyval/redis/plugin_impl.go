// Copyright (c) 2017 Cisco and/or its affiliates.
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

package redis

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/plugin"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/namsral/flag"
	"github.com/ligato/cn-infra/utils/safeclose"
)

// PluginID used in the Agent Core flavors
const PluginID core.PluginName = "RedisClient"

var defaultConfigFileName string

// Plugin implements Plugin interface therefore can be loaded with other plugins
type Plugin struct {
	LogFactory     logging.LogFactory
	ServiceLabel   *servicelabel.Plugin
	Connection     keyval.KvBytesPlugin
	ConfigFileName string
	Skeleton 	   *plugin.Skeleton
	logging.Logger
}

// Init is called on plugin startup. It establishes the connection to redis.
func (p *Plugin) Init() error {
	// Init logger
	var err error
	p.Logger, err = p.LogFactory.NewLogger(string(PluginID))
	if err != nil {
		return err
	}

	cfg, err := p.retrieveConfig()
	if err != nil {
		return err
	}
	client, err := CreateClient(cfg)
	if err != nil {
		return err
	}

	connection, err := NewBytesConnection(client, p.Logger)
	if err != nil {
		return err
	}

	skeleton := plugin.NewSkeleton(string(PluginID), p.LogFactory, p.ServiceLabel, connection)
	p.Skeleton = skeleton
	return p.Skeleton.Init()
}

// Close resources
func (p *Plugin) Close() error {
	_, err := safeclose.CloseAll(p.Skeleton, p.Connection)
	return err
}

func init() {
	flag.StringVar(&defaultConfigFileName, "redis-config", "",
		"Location of Redis configuration file; Can also be set via environment variable REDIS_CONFIG")
}

func (p *Plugin) retrieveConfig() (cfg interface{}, err error) {
	var configFile string
	if p.ConfigFileName != "" {
		configFile = p.ConfigFileName
	} else if defaultConfigFileName != "" {
		configFile = defaultConfigFileName
	}

	if configFile != "" {
		cfg, err = LoadConfig(configFile)
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}
