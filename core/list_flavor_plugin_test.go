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

package core

import (
	"testing"

	"github.com/onsi/gomega"
)

// Test01NoPluginsInFlavor checks that flavor that contains no fields
// that would implement Plugin interface (here we are missing Close())
// returns empty slice.
func TestListPlugins01NoPluginsInFlavor(t *testing.T) {
	gomega.RegisterTestingT(t)

	flavor := FlavorNoPlugin{}
	plugs := flavor.Plugins()
	t.Log("plugs ", plugs)
	gomega.Expect(plugs).To(gomega.BeNil())
}

// Test02OnePluginInFlavor checks that flavor that contains multiple fields
// but only one implements Plugin interface (another field is missing Close())
// returns slice with one particular plugin.
func TestListPlugins02OnePluginInFlavor(t *testing.T) {
	gomega.RegisterTestingT(t)

	flavor := &FlavorOnePlugin{}
	plugs := flavor.Plugins()
	t.Log("plugs ", plugs)
	gomega.Expect(plugs).To(gomega.Equal([]*NamedPlugin{&NamedPlugin{
		PluginName("Plugin2"), &flavor.Plugin2}}))
}

// TestListPlugins03ConfusingDepPointerToPlugin checks that flavor that contains multiple fields
// but only one implements Plugin interface (another field is missing Close()
// but has dependency - pointer to a plugin) returns slice with one particular plugin.
func TestListPlugins03ConfusingDepPointerToPlugin(t *testing.T) {
	gomega.RegisterTestingT(t)

	flavor := &FlavorConfusingDepPointerToPlugin{}
	plugs := flavor.Plugins()
	t.Log("plugs ", plugs)
	gomega.Expect(plugs).To(gomega.Equal([]*NamedPlugin{
		{PluginName("Plugin1"), &flavor.Plugin1},
		{PluginName("Dep01"), flavor.Plugin2.Dep01}}))
}

// FlavorNoPlugin contains no plugins
type FlavorNoPlugin struct {
	Plugin1 MissignCloseMethod
	Plugin2 struct {
		Dep1B string
	}
}

// FlavorOnePlugin contains one plugin (another is missing Close method)
type FlavorOnePlugin struct {
	Plugin1 MissignCloseMethod
	Plugin2 DummyPlugin
}

// FlavorConfusingDepPointerToPlugin one plugin
// & another field that seems to be plugin (but it is not)
type FlavorConfusingDepPointerToPlugin struct {
	Plugin1 DummyPlugin
	Plugin2 MissignCloseMethodWithDepPointingToPlugin
}

// MissignCloseMethod implements just Init() but not Close() method
type MissignCloseMethod struct {
}

// MissignCloseMethodWithDepPointingToPlugin intentionally
// contains pointer to dependency
type MissignCloseMethodWithDepPointingToPlugin struct {
	Dep01 *DummyPlugin
}

// Init does nothing
func (*MissignCloseMethod) Init() error {
	return nil
}

// DummyPlugin just defines Init() Close() with empty method bodies
type DummyPlugin struct {
}

// Init does nothing
func (*DummyPlugin) Init() error {
	return nil
}

// Close does nothing
func (*DummyPlugin) Close() error {
	return nil
}

// Init does nothing
func (*MissignCloseMethodWithDepPointingToPlugin) Init() error {
	return nil
}

// Plugins list plugins in this flavor
func (f *FlavorNoPlugin) Plugins() []*NamedPlugin {
	return ListPluginsInFlavor(f)
}

// Plugins list plugins in this flavor
func (f *FlavorOnePlugin) Plugins() []*NamedPlugin {
	return ListPluginsInFlavor(f)
}

// Plugins list plugins in this flavor
func (f *FlavorConfusingDepPointerToPlugin) Plugins() []*NamedPlugin {
	f.Plugin2.Dep01 = &f.Plugin1

	return ListPluginsInFlavor(f)
}
