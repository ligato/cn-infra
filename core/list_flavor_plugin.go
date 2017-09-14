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
	"errors"
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

// ListPluginsInFlavor lists plugins in a Flavor.
// It extracts all plugins and returns them as a slice of NamedPlugins.
func ListPluginsInFlavor(flavor Flavor) (plugins []*NamedPlugin) {
	uniqueness := map[PluginName]Plugin{}
	l, err := listPluginsInFlavor(reflect.ValueOf(flavor), uniqueness)
	if err != nil {
		logroot.StandardLogger().Error("Invalid argument - it does not satisfy the Flavor interface")
	}
	return l
}

// listPluginsInFlavor lists plugins in a Flavor. If there are multiple
// instances of a given plugin type, only one plugin instance is listed.
// A Flavor is composed of multiple Flavor and Plugins. The composition
// is recursive: a component Flavor contains Plugin components and may
// contain Flavor components as well. The function recursively lists
// plugins contained in component Flavors.
//
// The function returns an error if the flavorValue argument does not
// satisfy the Flavor interface. All components in the argument flavorValue
// must satisfy either the Plugin or the Flavor interface. If they do not,
// an error is logged, but the function does not return an error.
// in the argument
func listPluginsInFlavor(flavorValue reflect.Value, uniqueness map[PluginName]Plugin) ([]*NamedPlugin, error) {
	var res []*NamedPlugin

	flavorType := flavorValue.Type()

	if flavorType.Kind() == reflect.Ptr {
		flavorType = flavorType.Elem()
	}

	if flavorValue.Kind() == reflect.Ptr {
		flavorValue = flavorValue.Elem()
	}

	if !flavorValue.IsValid() {
		return res, nil
	}

	if _, ok := flavorValue.Addr().Interface().(Flavor); !ok {
		return res, errors.New("does not satisfy the Flavor interface")
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
				_, found := uniqueness[PluginName(field.Name)]
				if !found {
					uniqueness[PluginName(field.Name)] = plug
					res = append(res, &NamedPlugin{PluginName: PluginName(field.Name), Plugin: plug})
				}
			} else {
				// try to inspect flavor structure recursively
				l, err := listPluginsInFlavor(fieldVal, uniqueness)
				if err != nil {
					logroot.StandardLogger().
						WithField("fieldName", field.Name).
						Error("Bad field: must satisfy either Plugin or Flavor interface")
				} else {
					res = append(res, l...)
				}
			}
		}
	}

	return res, nil
}

// fieldPlugin determines if a given field satisfies the Plugin interface.
// If yes, the plugin value is returned; if not, nil is returned.
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
