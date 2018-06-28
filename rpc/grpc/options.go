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

package grpc

import (
	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
)

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	for _, o := range opts {
		o(p)
	}

	deps := &p.Deps
	if deps.PluginName == "" {
		deps.PluginName = "GRPC"
	}
	if deps.Log == nil {
		deps.Log = logging.ForPlugin(deps.PluginName.String(), logrus.DefaultRegistry)
	}
	if deps.PluginConfig == nil {
		deps.PluginConfig = config.ForPlugin(deps.PluginName.String())
	}

	return p
}

// Option is a function that acts on a Plugin to inject Dependencies or configuration
type Option func(*Plugin)

// UseDeps injects a particular set of Dependencies
func UseDeps(deps Deps) Option {
	return func(p *Plugin) {
		d := &p.Deps
		d.PluginName = deps.PluginName
		d.Log = deps.Log
		d.PluginConfig = deps.PluginConfig
		d.HTTP = deps.HTTP
	}
}

// UseConf injects the Plugin's Configuration
func UseConf(conf Config) Option {
	return func(p *Plugin) {
		p.grpcCfg = &conf
	}
}
