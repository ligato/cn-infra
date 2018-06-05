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

package grpc

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"testing"
)

func Test01DefaultWiring(t *testing.T) {
	plugin := &Plugin{}
	err := plugin.Wire(plugin.DefaultWiring(true))
	if err != nil {
		t.Errorf("plugin.Wire returned error %s", err)
	}
	if plugin.PluginName != core.PluginName(defaultName) {
		t.Errorf("Incorrect PluginName, expected %s, actual: %s", defaultName, string(plugin.PluginName))
	}
	if plugin.Log == nil {
		t.Errorf("Incorrect Log, expected non-nil, got nil")
	}
	if plugin.PluginConfig == nil {
		t.Errorf("Incorrect PluginConfig, expected non-nil got nil")
	}
}

func Test02WithNameDefault(t *testing.T) {
	plugin := &Plugin{}
	err := plugin.Wire(WithName(true))
	if err != nil {
		t.Errorf("plugin.Wire returned error %s", err)
	}
	if plugin.PluginName != core.PluginName(defaultName) {
		t.Errorf("Incorrect PluginName, expected %s, actual: %s", defaultName, string(plugin.PluginName))
	}

	if plugin.Log != nil {
		t.Errorf("Incorrect Log, expected nil, actual non-nil")
	}

	if plugin.PluginConfig != nil {
		t.Errorf("Incorrect PluginConfig, expected nil actual non-nil")
	}
}

func Test03WithNameNonDefault(t *testing.T) {
	plugin := &Plugin{}
	name := "foo"
	err := plugin.Wire(WithName(true, name))
	if err != nil {
		t.Errorf("plugin.Wire returned error %s", err)
	}
	if plugin.PluginName != core.PluginName(name) {
		t.Errorf("Incorrect PluginName, expected %s, actual: %s", name, string(plugin.PluginName))
	}
	if plugin.Log != nil {
		t.Errorf("Incorrect Log, expected nil, actual non-nil")
	}
	if plugin.PluginConfig != nil {
		t.Errorf("Incorrect PluginConfig, expected nil actual non-nil")
	}
}

func Test04WithLogNonDefault(t *testing.T) {
	plugin := &Plugin{}
	log := logging.ForPlugin(defaultName, logrus.NewLogRegistry())
	err := plugin.Wire(WithLog(true, log))
	if err != nil {
		t.Errorf("plugin.Wire returned error %s", err)
	}
	if plugin.Log != log {
		t.Errorf("Incorrect Log, expected %s, actual %s", log, plugin.Log)
	}
	if plugin.PluginConfig != nil {
		t.Errorf("Incorrect PluginConfig, expected nil actual non-nil")
	}
}

func Test05NilWiring(t *testing.T) {
	plugin := &Plugin{}
	err := plugin.Wire(nil)
	if err != nil {
		t.Errorf("plugin.Wire returned error %s", err)
	}
	if plugin.PluginName != core.PluginName(defaultName) {
		t.Errorf("Incorrect PluginName, expected %s, actual: %s", defaultName, string(plugin.PluginName))
	}
	if plugin.Log == nil {
		t.Errorf("Incorrect Log, expected non-nil, got nil")
	}
	if plugin.PluginConfig == nil {
		t.Errorf("Incorrect PluginConfig, expected non-nil got nil")
	}
}

func Test06DefaultWiringOverwriteTrue(t *testing.T) {
	plugin := &Plugin{}
	name := "foo"

	err := plugin.Wire(WithName(true, name))
	if err != nil {
		t.Errorf("plugin.Wire returned error %s", err)
	}

	err = plugin.Wire(DefaultWiring(false))
	if err != nil {
		t.Errorf("plugin.Wire returned error %s", err)
	}
	if plugin.PluginName != core.PluginName(name) {
		t.Errorf("Incorrect PluginName, expected %s, actual: %s", name, string(plugin.PluginName))
	}
	if plugin.Log == nil {
		t.Errorf("Incorrect Log, expected non-nil, actual nil")
	}
	if plugin.PluginConfig == nil {
		t.Errorf("Incorrect PluginConfig, expected nil actual non-nil")
	}

}
