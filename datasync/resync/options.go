package resync

import (
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
		p.Deps.PluginName = "resync"
	}
	if p.Deps.Log == nil {
		p.Deps.Log = logging.ForPlugin(p.Deps.PluginName.String(), logrus.DefaultRegistry)
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
	}
}
