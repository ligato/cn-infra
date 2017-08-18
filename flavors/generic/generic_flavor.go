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

package generic

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/httpmux"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/logging/logmanager"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/health/probe"
)

// FlavorGeneric glues together multiple plugins that are useful for almost every micro-service
type FlavorGeneric struct {
	Logrus       logrus.Plugin
	ServiceLabel servicelabel.Plugin
	Probe 		 probe.Plugin
	HTTP         httpmux.Plugin
	LogManager   logmanager.Plugin
	StatusCheck  statuscheck.Plugin

	injected bool
}

// Inject sets object references
func (f *FlavorGeneric) Inject() error {
	if f.injected {
		return nil
	}

	f.Probe.LogFactory = &f.Logrus
	f.HTTP.LogFactory = &f.Logrus
	f.HTTP.HTTPport = &f.Probe.HTTPPort
	//TODO f.HTTPProbe.Logger = f.Logrus.LoggerWithPrefix(f.PluginName(&f.HTTPProbe))
	//TODO f.HTTPProbe.Config = f.Config.ConfigWithPrefix(f.PluginName(&f.HTTPProbe))
	f.LogManager.ManagedLoggers = &f.Logrus
	f.LogManager.HTTP = &f.HTTP
	f.StatusCheck.LogFactory = &f.Logrus
	f.StatusCheck.Probe = &f.Probe.ProbePort
	f.StatusCheck.HTTP = &f.HTTP
	//f.StatusCheck.Transport todo inject transport

	f.injected = true

	return nil
}

// Plugins combines all Plugins in flavor to the list
func (f *FlavorGeneric) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
