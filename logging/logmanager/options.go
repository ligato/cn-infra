package logmanager

import (
	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
)

// DefaultPlugin is a default instance of Plugin.
var DefaultPlugin = NewPlugin()

// NewPlugin creates a new Plugin with the provided Options.
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

// Option is a function that acts on a Plugin to inject some settings.
type Option func(*Plugin)

// UseDeps returns Option which injects a particular set of dependencies.
func UseDeps(deps Deps) Option {
	return func(p *Plugin) {
		p.Deps = deps
	}
}

// UseConf returns Option which injects a particular configuration.
func UseConf(conf Conf) Option {
	return func(p *Plugin) {
		p.Conf = &conf
	}
}
