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

package kvdbsync

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/servicelabel"
)

// NewPlugin creates a new Plugin with the provided Options.
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	for _, o := range opts {
		o(p)
	}

	if p.Deps.PluginName == "" {
		prefix := "kvdb"
		if p.Deps.KvPlugin != nil {
			if kvdb, ok := p.Deps.KvPlugin.(core.PluginNamed); ok {
				prefix = kvdb.Name()
			}
		}
		p.Deps.PluginName = core.PluginName(prefix + "-datasync")
	}
	if p.Deps.Log == nil {
		p.Deps.Log = logging.ForPlugin(p.Deps.PluginName.String(), logrus.DefaultRegistry)
	}
	if p.Deps.ServiceLabel == nil {
		p.Deps.ServiceLabel = servicelabel.DefaultPlugin
	}

	return p
}

// Option is a function that acts on a Plugin to inject some settings.
type Option func(*Plugin)

// UseDeps returns Option which injects a particular set of dependencies.
func UseDeps(deps Deps) Option {
	return func(p *Plugin) {
		p.Deps = deps
	}
}
