package grpcsync

import (
	"github.com/ligato/cn-infra/rpc/grpc"
)

// NewPlugin creates a new Plugin with the provided Options.
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	for _, o := range opts {
		o(p)
	}

	if p.Deps.GRPC == nil {
		p.Deps.GRPC = grpc.DefaultPlugin
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
