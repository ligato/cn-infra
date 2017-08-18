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
	"github.com/ligato/cn-infra/logging"
)

// FlavorGeneric glues together multiple plugins that are useful for almost every micro-service
type FlavorGeneric struct {
	Logrus       logrus.Plugin
	ServiceLabel servicelabel.Plugin
	HealthProbe  probe.Plugin
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

	f.HTTP.Log = logging.NewPluginLogger(core.PluginNameOfFlavor(&f.HTTP, f), f.Logrus)
	//TODO f.HTTPProbe.Config = config.NewPluginConfig(f.PluginName(&f.HTTPProbe))
	f.LogManager.ManagedLoggers = &f.Logrus
	f.LogManager.HTTP = &f.HTTP
	f.StatusCheck.Log = logging.NewPluginLogger(core.PluginNameOfFlavor(&f.StatusCheck, f), f.Logrus)
	f.HealthProbe.Log = logging.NewPluginLogger(core.PluginNameOfFlavor(&f.HealthProbe, f), f.Logrus)
	f.HealthProbe.HTTP = &f.HTTP
	//f.HealthProbe.Transport todo inject local transport

	f.injected = true

	return nil
}

// Plugins combines all Plugins in flavor to the list
func (f *FlavorGeneric) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
