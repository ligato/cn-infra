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
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/persisted/dbsync"
	"github.com/ligato/cn-infra/db/keyval/plugin"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/statuscheck"
	"github.com/ligato/cn-infra/utils/config"
	"github.com/namsral/flag"
)

const (
	// PluginID used in the Agent Core flavors
	PluginID core.PluginName = "EtcdClient"
	// healthCheckProbeKey is a key used to probe Etcd state
	healthCheckProbeKey string = "/probe-etcd-connection"
)

// Plugin implements Plugin interface therefore can be loaded with other plugins
type Plugin struct {
	Transport      datasync.TransportAdapter
	LogFactory     logging.LogFactory
	ServiceLabel   *servicelabel.Plugin
	StatusCheck    *statuscheck.Plugin
	ConfigFileName string
	*plugin.Skeleton
}

var defaultConfigFileName string

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

// Init is called at plugin startup. The connection to etcd is established.
func (p *Plugin) Init() error {
	cfg, err := p.retrieveConfig()
	if err != nil {
		return err
	}

	// Init skeleton
	skeleton := plugin.NewSkeleton(string(PluginID),
		p.LogFactory,
		p.ServiceLabel,
		func(log logging.Logger) (plugin.Connection, error) {
			etcdConfig, err := ConfigToClientv3(cfg)
			if err != nil {
				return nil, err
			}
			return NewEtcdConnectionWithBytes(*etcdConfig, log)
		},
	)
	p.Skeleton = skeleton
	err = p.Skeleton.Init()
	if err != nil {
		return err
	}

	// Init ETCD transport
	p.Transport, err = p.InitTransport(p.Skeleton.Logger)
	if err != nil {
		return err
	}

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

// InitTransport initializes ETCD transport adapter which then can be injected to other plugins
func (p *Plugin) InitTransport(logger logging.Logger) (datasync.TransportAdapter, error) {
	cfg, err := p.retrieveConfig()
	if err != nil {
		return nil, err
	}
	etcdConfig, err := ConfigToClientv3(cfg)
	if err != nil {
		return nil, err
	}
	connection, err := NewEtcdConnectionWithBytes(*etcdConfig, logger)
	if err != nil {
		return nil, err
	}
	broker := connection.NewBroker(p.ServiceLabel.GetAgentPrefix())
	watcher := connection.NewWatcher(p.ServiceLabel.GetAgentPrefix())
	return dbsync.NewAdapter(string(PluginID), broker, watcher), nil
}
