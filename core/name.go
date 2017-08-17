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
	"reflect"

	"github.com/ligato/cn-infra/utils/structs"
)

// PluginName is a part of the plugin's API and it is supposed
// to be defined as a publicly accessible string constant.
// It is used to obtain the appropriate instance of the registry
// (there are multiple instances).
type PluginName string

// NamedPlugin represents a Plugin with a name
type NamedPlugin struct {
	PluginName
	Plugin
}

// String returns the PluginName
func (np *NamedPlugin) String() string {
	return string(np.PluginName)
}

// PluginNameOfFlavor tries to use reflection to find field name
// by traversing all field pointers of a structure instance
//
// Example usage to prepare logger name:
//
//    flavor.Logrus.NewLogger(PluginNameOfFlavor(&flavor.ETCD, flavor))
//
func PluginNameOfFlavor(ptrToPluginInFlavor interface{}, flavor Flavor) (
	name PluginName, found bool) {

	field, found := structs.FindField(flavor, ptrToPluginInFlavor)
	return pluginName(field), found
}

func pluginName(field *reflect.StructField) PluginName {
	if field == nil {
		return PluginName("")
	}
	return PluginName(field.Name)
}
