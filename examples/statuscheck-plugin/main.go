package main

import (
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/flavors/connectors"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/utils/safeclose"
)

// *************************************************************************
// This example demonstrates the usage of StatusReader API
// ETCD plugin is monitored by status check plugin.
// ExamplePlugin periodically prints the status.
// ************************************************************************/

func main() {
	// Init close channel used to stop the example.
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExamplePlugin, ETCDPlugin & FlavorLocal (reused cn-infra plugins).
	agent := local.NewAgent(local.WithPlugins(func(flavor *local.FlavorLocal) []*core.NamedPlugin {
		etcdPlug := &etcdv3.Plugin{}
		etcdDataSync := &kvdbsync.Plugin{}
		resyncOrch := &resync.Plugin{}

		etcdPlug.Deps.PluginInfraDeps = *flavor.InfraDeps("etcdv3", local.WithConf())
		resyncOrch.Deps.PluginLogDeps = *flavor.LogDeps("etcdv3-resync")
		connectors.InjectKVDBSync(etcdDataSync, etcdPlug, etcdPlug.PluginName, flavor, resyncOrch)

		examplePlug := &ExamplePlugin{closeChannel: exampleFinished}
		examplePlug.Deps.PluginInfraDeps = *flavor.InfraDeps("statuscheck-example")
		examplePlug.Deps.StatusMonitor = &flavor.StatusCheck // Inject status check

		return []*core.NamedPlugin{
			{etcdPlug.PluginName, etcdPlug},
			{etcdDataSync.PluginName, etcdDataSync},
			{resyncOrch.PluginName, resyncOrch},
			{examplePlug.PluginName, examplePlug}}
	}))
	core.EventLoopWithInterrupt(agent, nil)
}

// ExamplePlugin demonstrates the usage of datasync API.
type ExamplePlugin struct {
	Deps

	// Fields below are used to properly finish the example.
	closeChannel chan struct{}
}

// Init starts the consumer.
func (plugin *ExamplePlugin) Init() error {
	return nil
}

// AfterInit starts the publisher and prepares for the shutdown.
func (plugin *ExamplePlugin) AfterInit() error {

	go plugin.checkStatus(plugin.closeChannel)

	return nil
}

// checkStatus periodically prints status of plugins that publish their state
// to status check plugin
func (plugin *ExamplePlugin) checkStatus(closeCh chan struct{}) {
	for {
		select {
		case <-closeCh:
			plugin.Log.Info("Closing")
			return
		case <-time.After(1 * time.Second):
			status := plugin.StatusMonitor.GetAllPluginStatus()
			for k, v := range status {
				plugin.Log.Infof("Status[%v] = %v", k, v)
			}
		}
	}
}

// Close shutdowns the consumer and channels used to propagate data resync and data change events.
func (plugin *ExamplePlugin) Close() error {
	safeclose.CloseAll(plugin.closeChannel)
	return nil
}
