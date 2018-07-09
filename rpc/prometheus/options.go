package prometheus

import (
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/rpc/rest"
)

// DefaultPlugin is default instance of Plugin
var DefaultPlugin = NewPlugin()

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	for _, o := range opts {
		o(p)
	}

	deps := &p.Deps
	if deps.PluginName == "" {
		deps.PluginName = "grpc"
	}
	if deps.Log == nil {
		deps.Log = logging.ForPlugin(deps.PluginName.String(), logrus.DefaultRegistry)
	}
	if deps.HTTP == nil {
		deps.HTTP = rest.DefaultPlugin
	}
	/*if deps.PluginConfig == nil {
		deps.PluginConfig = config.ForPlugin(deps.PluginName.String())
	}*/

	return p
}

// Option is a function that acts on a Plugin to inject Dependencies or configuration
type Option func(*Plugin)

// UseDeps injects a particular set of Dependencies
func UseDeps(deps Deps) Option {
	return func(p *Plugin) {
		d := &p.Deps
		d.PluginName = deps.PluginName
		d.Log = deps.Log
		//d.PluginConfig = deps.PluginConfig
		d.HTTP = deps.HTTP
	}
}

// UseConf injects the Plugin's Configuration
/*func UseConf(conf Config) Option {
	return func(p *Plugin) {
		p.Config = &conf
	}
}
*/
