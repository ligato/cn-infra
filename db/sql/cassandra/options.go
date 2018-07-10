package cassandra

import (
	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/health/statuscheck"
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

	deps := &p.Deps
	if deps.PluginName == "" {
		deps.PluginName = "cassandra"
	}
	if deps.Log == nil {
		deps.Log = logging.ForPlugin(deps.PluginName.String(), logrus.DefaultRegistry)
	}
	if deps.PluginConfig == nil {
		deps.PluginConfig = config.ForPlugin(deps.PluginName.String())
	}
	if deps.StatusCheck == nil {
		deps.StatusCheck = statuscheck.DefaultPlugin
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
		p.Deps.PluginConfig = deps.PluginConfig
		p.Deps.StatusCheck = deps.StatusCheck
	}
}
