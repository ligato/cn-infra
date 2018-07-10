package servicelabel

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
