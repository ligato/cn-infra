package grpcsync

import (
	"github.com/ligato/cn-infra/rpc/grpc"
)

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	for _, o := range opts {
		o(p)
	}

	deps := &p.Deps
	if deps.GRPC == nil {
		deps.GRPC = grpc.DefaultPlugin
	}

	return p
}

// Option is a function that acts on a Plugin to inject Dependencies or configuration
type Option func(*Plugin)

// UseDeps injects a particular set of Dependencies
func UseDeps(deps Deps) Option {
	return func(p *Plugin) {
		d := &p.Deps
		d.GRPC = deps.GRPC
	}
}
