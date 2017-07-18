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
	log "github.com/ligato/cn-infra/logging/logrus"
	"reflect"
	"strings"
)

// ListPluginsInFlavor uses very simple reflection to traverse top level fields of Flavor structure.
// Each field is entry in response map. The key of the map is the name of the field.
func ListPluginsInFlavor(flavor interface{} /*pointer*/) (plugins map[PluginName]Plugin) {
	plugins = map[PluginName]Plugin{}
	listPluginsInFlavor(reflect.ValueOf(flavor), plugins)

	return plugins
}

// listPluginsInFlavor checks every field and tries to cast it to Plugin or inspect it's type recursively
func listPluginsInFlavor(flavorValue reflect.Value, plugins map[PluginName]Plugin) {
	flavorType := flavorValue.Type()
	log.WithField("flavorType", flavorType).Debug("ListPluginsInFlavor")

	if flavorType.Kind() == reflect.Ptr || flavorType.Kind() == reflect.Ptr {
		flavorType = flavorType.Elem()
	}

	if flavorValue.Kind() == reflect.Ptr || flavorValue.Kind() == reflect.Ptr {
		flavorValue = flavorValue.Elem()
	}

	if !flavorValue.IsValid() {
		log.WithField("flavorType", flavorType).Debug("invalid")
		return
	}

	pluginType := reflect.TypeOf((*Plugin)(nil)).Elem()

	if flavorType.Kind() == reflect.Struct {
		numField := flavorType.NumField()
		for i := 0; i < numField; i++ {
			field := flavorType.Field(i)
			exported := field.Name != "" && strings.ToUpper(string(field.Name[:1]))[0] == field.Name[0]
			if !exported {
				log.WithField("fieldName", field.Name).Debug("Unexported field")
				continue
			}

			fieldVal := flavorValue.Field(i)
			plug := fieldPlugin(field, fieldVal, pluginType)
			if plug != nil {
				plugins[PluginName(field.Name)] = plug
				log.WithField("fieldName", field.Name).Debug("Found plugin ", field.Type)
			} else {
				// try to inspect flavor structure recursively
				listPluginsInFlavor(fieldVal, plugins)
			}
		}
	}
}

// fieldPlugin tries to cast to Plugin
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
			log.WithField("fieldName", field.Name).Debug("Field is nil ", pluginType)
		} else if plug, ok := fieldVal.Interface().(Plugin); ok {
			return plug
		}

	}
	return nil
}

// Named translates map of plugins to slice of named plugins
func Named(plugins map[PluginName]Plugin) (outPlugins []*NamedPlugin) {
	outPlugins = make([]*NamedPlugin, len(plugins))
	i := 0
	for plugName, plug := range plugins {
		outPlugins[i] = &NamedPlugin{PluginName: plugName, Plugin: plug}
		i++
	}

	return outPlugins
}
