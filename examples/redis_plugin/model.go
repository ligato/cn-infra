package main

import (
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/db/keyval/redis"
	"github.com/ligato/cn-infra/flavors/local"
)

// ExampleFlavor is a set of plugins required for the redis example.
type ExampleFlavor struct {
	// Local flavor to access to Infra (logger, service label, status check)
	*local.FlavorLocal

	// Redis plugin
	Redis         redis.Plugin
	RedisDataSync kvdbsync.Plugin

	ResyncOrch resync.Plugin

	// Example plugin
	RedisExample ExamplePlugin

	// For example purposes, use channel when the example is finished
	closeChan *chan struct{}
}

// Deps is a helper struct which is grouping all dependencies injected to the plugin
type Deps struct {
	local.PluginLogDeps // injected
}
