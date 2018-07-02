package logmanager

import (
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
)

var DefaultPlugin = NewPlugin()

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	for _, o := range opts {
		o(p)
	}

	deps := &p.Deps
	if deps.PluginName == "" {
		deps.PluginName = "LogManager"
	}
	if deps.Log == nil {
		deps.Log = logging.ForPlugin(deps.PluginName.String(), logrus.DefaultRegistry)
	}

	return p
}

// Option is a function that acts on a Plugin to inject Dependencies or configuration
type Option func(*Plugin)

// UseDeps injects a particular set of Dependencies
func UseDeps(deps Deps) Option {
	return func(p *Plugin) {
		p.Deps.PluginName = deps.PluginName
		p.Deps.Log = deps.Log
		p.Deps.HTTP = deps.HTTP
		p.Deps.LogRegistry = deps.LogRegistry
	}
}

// UseConf injects the Plugin's Configuration
func UseConf(conf Conf) Option {
	return func(p *Plugin) {
		p.Conf = &conf
	}
}
