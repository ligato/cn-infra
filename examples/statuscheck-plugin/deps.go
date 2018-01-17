package main

import (
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/health/statuscheck"
)

// Deps lists dependencies of ExamplePlugin.
type Deps struct {
	local.PluginInfraDeps // injected
	StatusMonitor         statuscheck.StatusReader
}
