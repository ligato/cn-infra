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

// PluginName is a part of the plugin's API and it is supposed
// to be defined as a publicly accessible string constant.
// It is used to obtain the appropriate instance of the registry
// (there are multiple instances).
type PluginName string

// String returns the PluginName.
func (name PluginName) String() string {
	return string(name)
}

// NamedPlugin represents a Plugin with a name.
type NamedPlugin struct {
	PluginName
	Plugin
}

// String returns the PluginName.
func (np *NamedPlugin) String() string {
	return string(np.PluginName)
}

// Name returns the name of the plugin as a string
func (np *NamedPlugin) Name() string {
	return string(np.PluginName)
}

// PluginNamed is plugin with Name
type PluginNamed interface {
	Plugin
	Name() string
}

// NamePlugin returns NamedPlugin
func NamePlugin(name string, plugin Plugin) *NamedPlugin {
	return &NamedPlugin{
		PluginName: PluginName(name),
		Plugin:     plugin,
	}
}
