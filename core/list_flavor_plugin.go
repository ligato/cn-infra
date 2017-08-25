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

	"github.com/ligato/cn-infra/logging/logroot"
)

// Flavor is structure that contains a particular combination of plugins
// (fields of plugins)
type Flavor interface {
	// Plugins returns list of plugins.
	// Name of the plugin is supposed to be related to field name of Flavor struct
	Plugins() []*NamedPlugin
}

// ListPluginsInFlavor uses reflection to traverse top level fields of Flavor structure.
// It extracts all plugins and returns them as a slice of NamedPlugins.
func ListPluginsInFlavor(flavor Flavor) (plugins []*NamedPlugin) {
	pluginMap := map[PluginName]Plugin{}
	listPluginsInFlavor(reflect.ValueOf(flavor), pluginMap)

	plugins = []*NamedPlugin{}
	for plugName, plug := range pluginMap {
		if plug != nil {
			plugins = append(plugins, &NamedPlugin{PluginName: plugName, Plugin: plug})
		}
	}
	return plugins
}

// listPluginsInFlavor checks every field and tries to cast it to Plugin or inspect its type recursively.
func listPluginsInFlavor(flavorValue reflect.Value, pluginMap map[PluginName]Plugin) {
	flavorType := flavorValue.Type()

	if flavorType.Kind() == reflect.Ptr {
		flavorType = flavorType.Elem()
	}

	if flavorValue.Kind() == reflect.Ptr {
		flavorValue = flavorValue.Elem()
	}

	if !flavorValue.IsValid() {
		return
	}

	pluginType := reflect.TypeOf((*Plugin)(nil)).Elem()

	if flavorType.Kind() == reflect.Struct {
		numField := flavorType.NumField()
		for i := 0; i < numField; i++ {
			field := flavorType.Field(i)

			exported := field.PkgPath == "" // PkgPath is empty for exported fields
			if !exported {
				continue
			}

			fieldVal := flavorValue.Field(i)
			plug := fieldPlugin(field, fieldVal, pluginType)
			if plug != nil {
				pluginMap[PluginName(field.Name)] = plug
			} else {
				// try to inspect flavor structure recursively
				listPluginsInFlavor(fieldVal, pluginMap)
			}
		}
	}
}

// fieldPlugin tries to cast given field to Plugin
func fieldPlugin(field reflect.StructField, fieldVal reflect.Value, pluginType reflect.Type) Plugin {
	switch fieldVal.Kind() {
	case reflect.Struct:
		ptrType := reflect.PtrTo(fieldVal.Type())
		if ptrType.Implements(pluginType) && fieldVal.CanAddr() {
			if plug, ok := fieldVal.Addr().Interface().(Plugin); ok {
				return plug
			}
		}
	case reflect.Ptr, reflect.Interface:
		if fieldVal.IsNil() {
			logroot.StandardLogger().WithField("fieldName", field.Name).Debug("Field is nil ", pluginType)
		} else if plug, ok := fieldVal.Interface().(Plugin); ok {
			return plug
		}

	}
	return nil
}
