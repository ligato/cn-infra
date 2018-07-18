package infra

// Plugin interface defines plugin's basic life-cycle methods.
type Plugin interface {

	// Init is called in the agent`s startup phase.
	Init()

	// Close is called in the agent`s cleanup phase.
	Close() error

	// String returns unique name of the plugin.
	String() string
}

// PostInit interface defines an optional method for plugins with complex initialization.
type PostInit interface {
	// AfterInit is called once Init() of all plugins have returned without error.
	AfterInit() error
}
