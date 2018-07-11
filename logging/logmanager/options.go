package logmanager

// DefaultPlugin is a default instance of Plugin.
var DefaultPlugin = NewPlugin()

// NewPlugin creates a new Plugin with the provided Options.
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	for _, o := range opts {
		o(p)
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
