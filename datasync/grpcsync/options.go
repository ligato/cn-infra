package grpcsync

import (
	"github.com/ligato/cn-infra/rpc/grpc"
)

// DefaultPlugin is a default instance of Plugin.
var DefaultPlugin = *NewPlugin()

// NewPlugin creates a new Plugin with the provided Options.
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	p.GRPC = &grpc.DefaultPlugin

	for _, o := range opts {
		o(p)
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
