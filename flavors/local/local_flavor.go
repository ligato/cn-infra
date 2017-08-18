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

package local

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/servicelabel"

	"github.com/ligato/cn-infra/statuscheck"
)

// FlavorLocal glues together very minimal subset of cn-infra plugins
// that can be embeddable inside different project without running
// any agent specific server.
type FlavorLocal struct {
	Logrus       logrus.Plugin
	ServiceLabel servicelabel.Plugin
	StatusCheck  statuscheck.Plugin

	injected bool
}

// Inject does nothing (it is here for potential later extensibility)
// Composite flavors embedding local flavor are supposed to call this
// method.
func (f *FlavorLocal) Inject() error {
	if f.injected {
		return nil
	}

	f.injected = true

	return nil
}

// Plugins combines all Plugins in flavor to the list
func (f *FlavorLocal) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
