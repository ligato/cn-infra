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
	"context"
	"os"
	"reflect"
	"syscall"
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging/logrus"
)

// Options specifies the Version, MaxStartupTime, and Plugin list for the Agent
type Options struct {
	Version        string
	MaxStartupTime time.Duration
	QuitSignals    []os.Signal
	DoneChan       chan struct{}

	Plugins []core.PluginNamed

	ctx context.Context
}

func newOptions(opts ...Option) Options {
	opt := Options{
		Version:        "dev",
		MaxStartupTime: time.Second * 15,
		QuitSignals: []os.Signal{
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGKILL,
		},
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

// Version returns an Option that sets the version of the Agent to the entered string
func Version(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}

// Context returns an Option that sets the context for the Agent
func Context(ctx context.Context) Option {
	return func(o *Options) {
		o.ctx = ctx
	}
}

// QuitSignals returns an Option that will set signals which stop Agent
func QuitSignals(sigs ...os.Signal) Option {
	return func(o *Options) {
		o.QuitSignals = sigs
	}
}

func DoneChan(ch chan struct{}) Option {
	return func(o *Options) {
		o.DoneChan = ch
	}
}

// Plugins creates an Option that adds a list of Plugins to the Agent's Plugin list
func Plugins(plugins ...core.PluginNamed) Option {
	return func(o *Options) {
		o.Plugins = append(o.Plugins, plugins...)
	}
}

// AllPlugins creates an Option that adds all of the nested
// plugins recursively to the Agent's plugin list.
func AllPlugins(plugins ...core.Plugin) Option {
	return func(o *Options) {
		uniqueness := map[core.Plugin]interface{}{}
		for _, plugin := range plugins {
			plugins, err := listPlugins(reflect.ValueOf(plugin), uniqueness)
			if err != nil {
				panic(err)
			}
			o.Plugins = append(o.Plugins, plugins...)
			typ := reflect.TypeOf(plugin)
			logrus.DefaultLogger().Infof("recursively found %d plugins inside %v", len(plugins), typ)
			for _, plug := range plugins {
				logrus.DefaultLogger().Debugf(" - plugin: %v %v", reflect.TypeOf(plug), plug)
			}
			p, ok := plugin.(core.PluginNamed)
			if !ok {
				p = core.NamePlugin(typ.String(), plugin)
			}
			o.Plugins = append(o.Plugins, p)
		}
	}
}

func listPlugins(val reflect.Value, uniqueness map[core.Plugin]interface{}) ([]core.PluginNamed, error) {
	origTyp := val.Type()
	typ := val.Type()

	logrus.DefaultLogger().Debugf("=> inspect plugin structure for: %v (%v) %v ", typ, typ.Kind(), val)

	if typ.Kind() == reflect.Interface {
		if val.IsNil() {
			logrus.DefaultLogger().Debugf(" - val is nil")
			return nil, nil
		}
		val = val.Elem()
		typ = val.Type()
		logrus.DefaultLogger().Debugf(" - interface to: %v %v", typ, val)
	}

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		//logrus.DefaultLogger().Debug(" - typ ptr kind: ", typ)
	}
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		//logrus.DefaultLogger().Debug(" - val ptr kind: ", val)
	}

	if !val.IsValid() {
		logrus.DefaultLogger().Debugf(" - val is invalid")
		return nil, nil
	}

	/*if _, ok := val.Addr().Interface().(core.Plugin); !ok {
		return res, errors.New("does not satisfy the Plugin interface")
	}*/

	if typ.Kind() != reflect.Struct {
		logrus.DefaultLogger().Debugf(" - is not a struct: %v %v", typ.Kind(), val.Kind())
		return nil, nil
	}

	var res []core.PluginNamed

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		exported := field.PkgPath == "" // PkgPath is empty for exported fields
		if !exported {
			continue
		}

		fieldVal := val.Field(i)

		logrus.DefaultLogger().Debugf(" - check %v field[%d]: %v %v", typ, i, field.Type, fieldVal)

		var fieldPlug core.PluginNamed
		plug, implementsPlugin := isFieldPlugin(field, fieldVal)
		if implementsPlugin {
			if plug == nil {
				logrus.DefaultLogger().WithField("fieldName", field.Name).
					Debug(" - found nil plugin")
				continue
			}

			_, found := uniqueness[plug]
			if found {
				logrus.DefaultLogger().WithField("fieldName", field.Name).
					Debugf(" - Found duplicate plugin: %v", field.Type)
				continue
			}

			uniqueness[plug] = nil
			p, ok := plug.(core.PluginNamed)
			if !ok {
				p = core.NamePlugin(field.Name, plug)
			}
			//res = append(res, p)
			fieldPlug = p

			logrus.DefaultLogger().WithField("fieldName", field.Name).
				Warnf(" - Found plugin: %v (%v)", p.Name(), field.Type)
		}

		// try to inspect plugin structure recursively
		l, err := listPlugins(fieldVal, uniqueness)
		if err != nil {
			logrus.DefaultLogger().WithField("fieldName", field.Name).
				Debug(" - Bad field: ", err)
			continue
		}

		logrus.DefaultLogger().Debugf(" - listed %v plugins from %v (%v)", len(l), field.Name, field.Type)
		res = append(res, l...)

		if fieldPlug != nil {
			res = append(res, fieldPlug)
		}
	}

	logrus.DefaultLogger().Debugf("-> found %v plugins in %v (%v)", len(res), typ, origTyp.Kind())

	return res, nil
}

var pluginType = reflect.TypeOf((*core.Plugin)(nil)).Elem()

func isFieldPlugin(field reflect.StructField, fieldVal reflect.Value) (core.Plugin, bool) {

	logrus.DefaultLogger().Debugf(" - is field plugin: %v (%v) %v", field.Type, fieldVal.Kind(), fieldVal)

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
				return nil, true
			}
			return plug, true
		} else {
			logrus.DefaultLogger().Debugf(" - does not implement Plugin: %v", field.Type.Implements(pluginType))
		}
		/*case reflect.Interface:
		fieldVal = fieldVal.Elem()
		if !fieldVal.IsValid() {
			return nil, true
		}
		if plug, ok := fieldVal.Interface().(core.Plugin); ok {
			if fieldVal.IsNil() {
				return nil, true
			}
			return plug, true
		} else {
			logrus.DefaultLogger().Debugf(" - does not implement Plugin: %v", fieldVal.Type())
		}*/
	}

	return nil, false
}
