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
	"github.com/ligato/cn-infra/db/keyval/plugin"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/utils/safeclose"
)

const (
	// healthCheckProbeKey is a key used to probe Redis state
	healthCheckProbeKey string = "probe-redis-connection"
)

// Plugin implements Plugin interface therefore can be loaded with other plugins
type Plugin struct {
	Deps
	*plugin.Skeleton
	disabled bool
}

// Deps is here to group injected dependencies of plugin
// to not mix with other plugin fields.
type Deps struct {
	local.PluginInfraDeps //inject
}

// Init is called on plugin startup. It establishes the connection to redis.
func (p *Plugin) Init() error {
	cfg, err := p.retrieveConfig()
	if err != nil {
		return err
	}
	if p.disabled {
		return nil
	}

	client, err := CreateClient(cfg)
	if err != nil {
		return err
	}

	if p.Skeleton == nil {
		con, err := NewBytesConnection(client, p.Log)
		if err != nil {
			return err
		}

		p.Skeleton = plugin.NewSkeleton(p.String(),
			p.ServiceLabel,
			con,
		)
	}
	err = p.Skeleton.Init()
	if err != nil {
		return err
	}

	return nil
}

// AfterInit is called by the Agent Core after all plugins have been initialized.
func (p *Plugin) AfterInit() error {
	if p.disabled {
		return nil
	}

	// Register for providing status reports (polling mode)
	if p.StatusCheck != nil {
		p.StatusCheck.Register(core.PluginName(p.String()), func() (statuscheck.PluginState, error) {
			_, _, err := p.Skeleton.NewBroker("/").GetValue(healthCheckProbeKey, nil)
			if err == nil {
				return statuscheck.OK, nil
			}
			return statuscheck.Error, err
		})
	} else {
		p.Log.Warnf("Unable to start status check for redis")
	}

	return nil
}

// Close resources
func (p *Plugin) Close() error {
	_, err := safeclose.CloseAll(p.Skeleton)
	return err
}

func (p *Plugin) retrieveConfig() (cfg interface{}, err error) {
	found, _ := p.PluginConfig.GetValue(&struct{}{})
	if !found {
		p.Log.Info("redis config not found ", p.PluginConfig.GetConfigName(), " - skip loading this plugin")
		p.disabled = true
		return nil, nil
	}
	configFile := p.PluginConfig.GetConfigName()
	if configFile != "" {
		cfg, err = LoadConfig(configFile)
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// String returns if set Deps.PluginName or "redis-client" otherwise
func (p *Plugin) String() string {
	if len(p.Deps.PluginName) == 0 {
		return "redis-client"
	}
	return string(p.Deps.PluginName)
}

// Disabled if the plugin was not found
func (p *Plugin) Disabled() (disabled bool) {
	return p.disabled
}
