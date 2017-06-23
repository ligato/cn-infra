package core

// PluginName is a part of the plugin's API and it is supposed
// to be defined as a publicly accessible string constant.
// It is used to obtain the appropriate instance of the registry
// (there are multiple instances).
type PluginName string

// NamedPlugin represents a Plugin with a name
type NamedPlugin struct {
	PluginName
	Plugin
}

// String prints the PluginName
func (np *NamedPlugin) String() string {
	return string(np.PluginName)
}

// Plugin interface is used to pass custom plugins instances to the agent
type Plugin interface {
	Init() error
	Close() error
}

// PostInit interface is used to pass custom plugins instances that support AfterInit() to the agent
type PostInit interface {
	// Is meant for
	AfterInit() error
}
