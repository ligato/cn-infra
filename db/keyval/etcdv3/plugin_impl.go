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

package etcdv3

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/plugin"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/utils/config"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/namsral/flag"
)

const (
	// PluginID used in the Agent Core flavors
	PluginID core.PluginName = "EtcdClient"
	// healthCheckProbeKey is a key used to probe Etcd state
	healthCheckProbeKey string = "/probe-etcd-connection"
)

var defaultConfigFileName string

// Plugin implements Plugin interface therefore can be loaded with other plugins
type Plugin struct {
	LogFactory logging.LogFactory
	logging.Logger

	ServiceLabel *servicelabel.Plugin
	StatusCheck  *statuscheck.Plugin
	Skeleton     *plugin.Skeleton

	Connection     keyval.KvBytesPlugin
	ConfigFileName string
}

// Init is called at plugin startup. The connection to etcd is established.
func (p *Plugin) Init() error {
	var err error
	// Init logger
	p.Logger, err = p.LogFactory.NewLogger(string(PluginID))
	if err != nil {
		return err
	}

	// Retrieve config
	cfg, err := p.retrieveConfig()
	if err != nil {
		return err
	}

	// Init connection
	etcdConfig, err := ConfigToClientv3(cfg)
	if err != nil {
		return err
	}
	connection, err := NewEtcdConnectionWithBytes(*etcdConfig, p.Logger)
	if err != nil {
		return err
	}

	// Init skeleton
	skeleton := plugin.NewSkeleton(string(PluginID),
		p.LogFactory,
		p.ServiceLabel,
		connection,
	)
	p.Skeleton = skeleton
	err = p.Skeleton.Init()
	if err != nil {
		return err
	}

	// Connection is passed to DbSync plugin to register transport
	p.Connection = connection

	// Register for providing status reports (polling mode)
	if p.StatusCheck != nil {
		p.StatusCheck.Register(PluginID, func() (statuscheck.PluginState, error) {
			_, _, err := p.Skeleton.NewBroker("/").GetValue(healthCheckProbeKey, nil)
			if err == nil {
				return statuscheck.OK, nil
			}
			return statuscheck.Error, err
		})
	} else {
		p.Skeleton.Logger.Warnf("Unable to start status check for etcd")
	}

	return nil
}

// Close resources
func (p *Plugin) Close() error {
	_, err := safeclose.CloseAll(p.Skeleton, p.Connection)
	return err
}

func init() {
	flag.StringVar(&defaultConfigFileName, "etcdv3-config", "", "Location of the Etcd configuration file; also set via 'ETCDV3_CONFIG' env variable.")
}

func (p *Plugin) retrieveConfig() (*Config, error) {
	cfg := &Config{}
	var configFile string
	if p.ConfigFileName != "" {
		configFile = p.ConfigFileName
	} else if defaultConfigFileName != "" {
		configFile = defaultConfigFileName
	}

	if configFile != "" {
		err := config.ParseConfigFromYamlFile(configFile, cfg)
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}
