package main

import (
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/flavors/local"
	log "github.com/ligato/cn-infra/logging/logroot"
	"github.com/namsral/flag"
)

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized. Agent is instantiated with generic plugin (Status check, and Log)
// and example plugin which demonstrates use of Redis flavor.
func main() {
	// Init close channel used to stop the example
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExampleFlavor (combination of ExamplePlugin & reused cn-infra plugins)
	flavor := ExampleFlavor{closeChan: &exampleFinished}
	agent := core.NewAgent(log.StandardLogger(), 15*time.Second, append(flavor.Plugins())...)
	core.EventLoopWithInterrupt(agent, exampleFinished)
}

// Redis flag to load config
func init() {
	flag.String("redis-config", "redis.conf",
		"Location of the redis configuration file")
}

// Inject sets object references
func (ef *ExampleFlavor) Inject() (allReadyInjected bool) {
	// Init local flavor
	if ef.FlavorLocal == nil {
		ef.FlavorLocal = &local.FlavorLocal{}
	}
	ef.FlavorLocal.Inject()
	ef.Redis.Deps.PluginInfraDeps = *ef.InfraDeps("redis")
	ef.ResyncOrch.Deps.PluginLogDeps = *ef.LogDeps("redis-resync")
	InjectKVDBSync(&ef.RedisDataSync, &ef.Redis, ef.Redis.PluginName, ef.FlavorLocal, &ef.ResyncOrch)
	ef.RedisExample.Deps.PluginLogDeps = *ef.FlavorLocal.LogDeps("redis-example")
	ef.RedisExample.closeChannel = ef.closeChan

	return true
}

// InjectKVDBSync helper to set object references
func InjectKVDBSync(dbsync *kvdbsync.Plugin,
	db keyval.KvProtoPlugin, dbPlugName core.PluginName, local *local.FlavorLocal, resync resync.Subscriber) {

	dbsync.Deps.PluginLogDeps = *local.LogDeps(string(dbPlugName) + "-datasync")
	dbsync.KvPlugin = db
	dbsync.ResyncOrch = resync
	if local != nil {
		dbsync.ServiceLabel = &local.ServiceLabel

		if local.StatusCheck.Transport == nil {
			local.StatusCheck.Transport = dbsync
		}
	}
}

// Plugins combines all Plugins in flavor to the list
func (ef *ExampleFlavor) Plugins() []*core.NamedPlugin {
	ef.Inject()
	return core.ListPluginsInFlavor(ef)
}

// ExamplePlugin to depict the use of Redis flavor
type ExamplePlugin struct {
	Deps // plugin dependencies are injected

	closeChannel *chan struct{}
}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() (err error) {
	return nil
}

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime
func (plugin *ExamplePlugin) Close() error {
	*plugin.closeChannel <- struct{}{}
	return nil
}
