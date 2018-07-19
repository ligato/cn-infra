package resync

import (
	"github.com/ligato/cn-infra/logging"
)

// DefaultPlugin is a default instance of Plugin.
var DefaultPlugin Plugin = *NewPlugin()

// NewPlugin creates a new Plugin with the provided Options.
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	p.PluginName = "resync"

	for _, o := range opts {
		o(p)
	}

	if p.Deps.Log == nil {
		p.Deps.Log = logging.ForPlugin(p.String())
	}

	return p
}

// Option is a function that acts on a Plugin to inject some settings.
type Option func(*Plugin)

// UseDeps returns Option that can inject custom dependencies.
func UseDeps(cb func(*Deps)) Option {
	return func(p *Plugin) {
		cb(&p.Deps)
	}
}
