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

package probe

import (
	"github.com/namsral/flag"
	"github.com/ligato/cn-infra/httpmux"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/utils/safeclose"
)

// Default port used for http and probing
const defaultPort = "9191"

// PluginID used in the Agent Core flavors
const PluginID core.PluginName = "Probe"

var (
	httpPort string
	probePort string
)

// init is here only for parsing program arguments
func init() {
	flag.StringVar(&httpPort, "http-port", defaultPort,
		"Listen port for the Agent's HTTPProbe server.")
	flag.StringVar(&probePort, "probe-http-port", defaultPort,
		"Listen probe port for the Agent's HTTPProbe server.")
}

// Plugin struct holds all plugin-related data
type Plugin struct {
	LogFactory logging.LogFactory
	logging.Logger

	HTTPPort    httpmux.HTTPPort
	ProbePort   httpmux.HTTPPort
	CustomProbe bool
}

// Init is the plugin entry point called by the Agent Core
func (p *Plugin) Init() error {
	var err error
	p.Logger, err = p.LogFactory.NewLogger(string(PluginID))
	if err != nil {
		return err
	}
	p.HTTPPort = httpmux.HTTPPort{
		Port: httpPort,
	}
	p.ProbePort = httpmux.HTTPPort{
		Port: probePort,
	}

	return nil
}

// Close is called by the Agent Core when it's time to clean up the plugin
func (p *Plugin) Close() error {
	_, err := safeclose.CloseAll(p.LogFactory)
	return err
}

