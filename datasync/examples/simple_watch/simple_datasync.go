package main

import (
	"context"

	"os"

	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/flavors/etcdkafka"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/utils/safeclose"
)

func main() {
	logroot.Logger().SetLevel(logging.DebugLevel)

	f := etcdkafka.Flavor{}
	//TODO inject PluginXY
	agent := core.NewAgent(logroot.Logger(), 15*time.Second, f.Plugins()...)

	err := core.EventLoopWithInterrupt(agent, nil)
	if err != nil {
		os.Exit(1)
	}
}

// ExamplePlugin does watching
type PluginXY struct {
	Watcher   datasync.Watcher //Injected
	Logger    logging.Logger
	ParentCtx context.Context

	dataChange chan datasync.ChangeEvent
	dataResync chan datasync.ResyncEvent
	cancel     context.CancelFunc
}

// Init initializes channels & go routine
func (plugin *PluginXY) Init() (err error) {
	// initialize channels & start go routins
	plugin.dataChange = make(chan datasync.ChangeEvent, 100)
	plugin.dataResync = make(chan datasync.ResyncEvent, 100)

	// initiate context & cancel function (to stop go routine)
	var ctx context.Context
	if plugin.ParentCtx == nil {
		ctx, plugin.cancel = context.WithCancel(context.Background())
	} else {
		ctx, plugin.cancel = context.WithCancel(plugin.ParentCtx)
	}

	go func() {
		for {
			select {
			case dataChangeEvent := <-plugin.dataChange:
				plugin.Logger.Debug(dataChangeEvent)
			case dataResyncEvent := <-plugin.dataResync:
				plugin.Logger.Debug(dataResyncEvent)
			case <-ctx.Done():
				// stop watching for notifications
				return
			}
		}
	}()

	return nil
}

// AfterInit starts watching the data
func (plugin *PluginXY) AfterInit() error {
	// subscribe plugin.channel for watching data (to really receive the data)
	plugin.Watcher.WatchData("watchingXY", plugin.dataChange, plugin.dataResync, "keysXY")

	return nil
}

// Close cancels the go routine and close channels
func (plugin *PluginXY) Close() error {
	// cancel watching the channels
	plugin.cancel()

	// close all resources / channels
	_, err := safeclose.CloseAll(plugin.dataChange, plugin.dataResync)
	return err
}
