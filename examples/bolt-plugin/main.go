package main

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/bolt"
	"github.com/ligato/cn-infra/flavors/local"
)

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized. Agent is instantiated with generic plugin (Status check, and Log)
// and example plugin which demonstrates use of Redis flavor.
func main() {
	// Init close channel used to stop the example
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExamplePlugin, BoltPlugin & FlavorLocal (reused cn-infra plugins).
	agent := local.NewAgent(local.WithPlugins(func(flavor *local.FlavorLocal) []*core.NamedPlugin {
		boltPlug := &bolt.Plugin{}
		boltPlug.Deps.PluginInfraDeps = *flavor.InfraDeps("bolt", local.WithConf())

		examplePlug := &ExamplePlugin{closeChannel: &exampleFinished}
		examplePlug.Deps.PluginLogDeps = *flavor.LogDeps("bolt-example")
		examplePlug.Deps.DB = boltPlug // Inject Bolt to example plugin.

		return []*core.NamedPlugin{
			{boltPlug.PluginName, boltPlug},
			{examplePlug.PluginName, examplePlug}}
	}))
	core.EventLoopWithInterrupt(agent, exampleFinished)
}

// ExamplePlugin to depict the use of Redis flavor
type ExamplePlugin struct {
	Deps // plugin dependencies are injected

	closeChannel *chan struct{}
}

// Deps is a helper struct which is grouping all dependencies injected to the plugin
type Deps struct {
	local.PluginLogDeps                      // injected
	DB                  keyval.KvProtoPlugin // injected
}

// Init is meant for registering the watcher
func (plugin *ExamplePlugin) Init() (err error) {
	//TODO plugin.Watcher.Watch()

	return nil
}

// AfterInit is meant to use DB if needed
func (plugin *ExamplePlugin) AfterInit() (err error) {
	db := plugin.DB.NewBroker(keyval.Root)

	// Store some data
	txn := db.NewTxn()
	txn.Put("/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop0", nil)
	txn.Put("/vnf-agent/agent_vpp_1/vpp/config/v1/interface/bvi_loop1", nil)
	txn.Put("/vnf-agent/agent_vpp_1/vpp/config/v1/bd/b1", nil)
	txn.Put("/vnf-agent/agent_vpp_1/vpp/config/v1/bd/b2", nil)
	txn.Put("/vnf-agent/agent_vpp_1/vpp/config/v2/bd/b1", nil)
	txn.Commit()

	// List keys
	plugin.Log.Info("List Bolt DB keys /vnf-agent/agent_vpp_1/vpp/config/")
	keys, err := db.ListKeys("/vnf-agent/agent_vpp_1/vpp/config/")
	if keys != nil {
		for {
			k, _, all := keys.GetNext()
			if all == true {
				break
			}
			plugin.Log.Infof("Key : %v", k)
		}
	}
	return nil
}

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime
func (plugin *ExamplePlugin) Close() error {
	*plugin.closeChannel <- struct{}{}
	return nil
}
