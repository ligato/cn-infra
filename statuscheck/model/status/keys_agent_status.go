package status

const (
	// StatusPrefix is the relative key prefix for the agent/plugin status.
	StatusPrefix = "check/status/v1/"
	// AgentStatusPrefix filters status of the agent (not all other plugins)
	AgentStatusPrefix = StatusPrefix + "agent"
)

// AgentStatusKey returns the key used in ETCD to store the operational status of the vpp agent instance.
func AgentStatusKey() string {
	return AgentStatusPrefix
}

// PluginStatusKey returns the key used in ETCD to store the operational status of the vpp agent plugin.
func PluginStatusKey(pluginLabel string) string {
	return StatusPrefix + "plugin/" + pluginLabel
}
