package logmanager

import (
	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
)

// DefaultPlugin is default instance of Plugin
var DefaultPlugin = NewPlugin()

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	for _, o := range opts {
		o(p)
	}

	if p.Deps.PluginName == "" {
		p.Deps.PluginName = "logs"
	}
	if p.Deps.Log == nil {
		p.Deps.Log = logging.ForPlugin(p.Deps.PluginName.String(), logrus.DefaultRegistry)
	}
	if p.Deps.PluginConfig == nil {
		p.Deps.PluginConfig = config.ForPlugin(p.Deps.PluginName.String())
	}
	if p.Deps.LogRegistry == nil {
		p.Deps.LogRegistry = logrus.DefaultRegistry
	}

	return p
}

// Option is a function that acts on a Plugin to inject Dependencies or configuration
type Option func(*Plugin)

// UseDeps injects a particular set of Dependencies
func UseDeps(deps Deps) Option {
	return func(p *Plugin) {
		p.Deps = deps
	}
}

// UseConf injects the Plugin's Configuration
func UseConf(conf Conf) Option {
	return func(p *Plugin) {
		p.Conf = &conf
	}
}
