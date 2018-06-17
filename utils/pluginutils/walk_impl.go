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

package pluginutils

import (
	"reflect"

	"fmt"

	"github.com/ligato/cn-infra/core"
)

const (
	invalidValueErrorString = "Invalid value: %s"
)

type WalkFunc func(core.Plugin) error

// Walk runs WalkFunc f on all descendants of
func Walk(plugin core.Plugin, f WalkFunc) error {
	err := f(plugin)
	if err != nil {
		return err
	}
	// Get and pluginValue
	pluginValue := reflect.ValueOf(plugin)

	// If the plugin value is a pointer, get a concrete value
	if pluginValue.Kind() == reflect.Ptr {
		pluginValue = pluginValue.Elem()
	}

	// Check for zero value etc with isValid()
	if !pluginValue.IsValid() {
		return fmt.Errorf(invalidValueErrorString, pluginValue)
	}

	// We need the type so we can recurse the *type* and get info about fields like their names, whether they are
	// exported etc
	pluginType := pluginValue.Type()

	// If the plugin isn't a Struct... what are we doing here? :)
	if pluginType.Kind() == reflect.Struct {
		// Iterate over the Fields in the Struct
		numField := pluginType.NumField()
		for i := 0; i < numField; i++ {
			field := pluginType.Field(i)

			// If its not exported, ignore
			// PkgPath is empty for exported fields because there is no restriction on which pkg can access them
			exported := field.PkgPath == ""
			if !exported {
				continue
			}

			// Now we see if any of the values in those fields are actually Plugins
			// Note, its not enough to inspect the types of those fields
			// The field type represents what is defined in the Struct, and the Struct
			// May have a non-plugin interface as its field type
			// But if the *value* in this particular Struct is a plugin, we need to know that
			fieldVal := pluginValue.Field(i)
			// If its a pointer, get the concrete value
			if fieldVal.Kind() == reflect.Ptr {
				fieldVal = fieldVal.Elem()
			}
			// Check to see if that concrete value is a core.Plugin and not nil
			// Note: Always check CanAddr() or a Panic can results
			if fieldVal.CanAddr() {
				pluginInterface := fieldVal.Addr().Interface()
				plug, ok := pluginInterface.(core.Plugin)
				if ok && plug != nil {
					return Walk(plug, f)
				}
			}
		}
	}
	return nil
}

type List struct {
	Plugins []core.Plugin
}

func (pl *List) WalkFunc(plugin core.Plugin) error {
	pl.Plugins = append(pl.Plugins, plugin)
	return nil
}
