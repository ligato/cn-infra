package wiring

import (
	"reflect"

	"github.com/ligato/cn-infra/core"
)

// ListUniquePlugins lists the Unique Plugins in a given Plugin
// Takes argument *core.Plugin
// Returns []*core.Plugin
func ListUniqueNamedPlugins(plugins ...core.Plugin) []*core.NamedPlugin {
	var res []*core.NamedPlugin
	uniqueness := map[core.Plugin] /*nil*/ interface{}{}
	for _, plugin := range plugins {

		// Get and pluginValue
		pluginValue := reflect.ValueOf(plugin)

		// If the plugin value is a pointer, get a concrete value
		if pluginValue.Kind() == reflect.Ptr {
			pluginValue = pluginValue.Elem()
		}

		// Check for zero value etc with isValid()
		if !pluginValue.IsValid() {
			return res
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
						// Check to see if the plugin is unique, ie we haven't seen it before
						_, found := uniqueness[plug]
						if !found {
							// Note that we have seen this plugin
							uniqueness[plug] = nil
							// Append it to the list
							res = append(res, &core.NamedPlugin{PluginName: core.PluginName(field.Name), Plugin: plug})
							l := ListUniqueNamedPlugins(plug)
							// Do a uniqueness check for the
							for _, np := range l {
								_, found := uniqueness[np.Plugin]
								if !found {
									uniqueness[np.Plugin] = nil
									res = append(res, np)
								}
							}

						}
					}
				}
			}

		}
	}
	return res
}
