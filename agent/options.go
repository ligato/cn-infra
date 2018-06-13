//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package agent

import (
	"errors"
	"reflect"
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging/logrus"
)

// Options specifies the Version, MaxStartupTime, and Plugin list for the Agent
type Options struct {
	Version        string
	MaxStartupTime time.Duration

	Plugins []core.PluginNamed
}

func newOptions(opts ...Option) Options {
	opt := Options{
		Version:        "dev",
		MaxStartupTime: time.Second * 15,
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// Option is a function that operates on an Agent's Option
type Option func(*Options)

// MaxStartupTime returns an Option that sets the MaxStartuptime option of the Agent
func MaxStartupTime(d time.Duration) Option {
	return func(o *Options) {
		o.MaxStartupTime = d
	}
}

// Version returns an Option that sets the version of the Agent to the string v
func Version(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}

// Plugins creates an Option that adds a list of Plugins to the Agent's Plugin list
func Plugins(plugins ...core.PluginNamed) Option {
	return func(o *Options) {
		o.Plugins = append(o.Plugins, plugins...)
	}
}

// Recursive creates an Option that adds all of the Descendants of a Plugin Recursively to the Agent's
// Plugin list
func Recursive(plugin core.Plugin) Option {
	return func(o *Options) {
		uniqueness := map[core.Plugin]interface{}{}
		plugins, err := listPlugins(reflect.ValueOf(plugin), uniqueness)
		if err != nil {
			panic(err)
		}
		o.Plugins = append(o.Plugins, plugins...)
		typ := reflect.TypeOf(plugin)
		logrus.DefaultLogger().Infof("recursively found %d plugins inside %v", len(plugins), typ)
		o.Plugins = append(o.Plugins, core.NamePlugin(typ.String(), plugin))
	}
}

func listPlugins(val reflect.Value, uniqueness map[core.Plugin]interface{}) ([]core.PluginNamed, error) {
	logrus.DefaultLogger().Debug("inspect plugin structure ", val.Type())

	var res []core.PluginNamed

	typ := val.Type()

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if !val.IsValid() {
		return res, nil
	}

	if _, ok := val.Addr().Interface().(core.Plugin); !ok {
		return res, errors.New("does not satisfy the Plugin interface")
	}

	pluginType := reflect.TypeOf((*core.Plugin)(nil)).Elem()

	if typ.Kind() == reflect.Struct {
		numField := typ.NumField()
		for i := 0; i < numField; i++ {
			field := typ.Field(i)

			exported := field.PkgPath == "" // PkgPath is empty for exported fields
			if !exported {
				continue
			}

			fieldVal := val.Field(i)
			plug, implementsPlugin := isFieldPlugin(field, fieldVal, pluginType)
			if implementsPlugin {
				if plug != nil {
					_, found := uniqueness[plug]
					if !found {
						uniqueness[plug] = nil
						res = append(res, core.NamePlugin(field.Name, plug))

						logrus.DefaultLogger().
							WithField("fieldName", field.Name).
							Debug("Found plugin ", field.Type)
					} else {
						logrus.DefaultLogger().
							WithField("fieldName", field.Name).
							Debug("Found plugin with non unique name")
					}
				} else {
					logrus.DefaultLogger().
						WithField("fieldName", field.Name).
						Debug("Found nil plugin")
				}
			} else {
				// try to inspect plugin structure recursively
				l, err := listPlugins(fieldVal, uniqueness)
				if err != nil {
					logrus.DefaultLogger().
						WithField("fieldName", field.Name).
						Error("Bad field: must satisfy Plugin: ", err)
				} else {
					res = append(res, l...)
				}
			}
		}
	}

	return res, nil
}

func isFieldPlugin(field reflect.StructField, fieldVal reflect.Value, pluginType reflect.Type) (
	plugin core.Plugin, implementsPlugin bool) {

	switch fieldVal.Kind() {
	case reflect.Struct:
		ptrType := reflect.PtrTo(fieldVal.Type())
		if ptrType.Implements(pluginType) {
			if fieldVal.CanAddr() {
				if plug, ok := fieldVal.Addr().Interface().(core.Plugin); ok {
					return plug, true
				}
			}
			return nil, true
		}
	case reflect.Ptr, reflect.Interface:
		if plug, ok := fieldVal.Interface().(core.Plugin); ok {
			if fieldVal.IsNil() {
				logrus.DefaultLogger().WithField("fieldName", field.Name).
					Debug("Field is nil ", pluginType)
				return nil, true
			}
			return plug, true
		}

	}
	return nil, false
}
