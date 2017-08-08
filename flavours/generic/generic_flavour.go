package generic

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/httpmux"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/servicelabel"

	"github.com/ligato/cn-infra/logging/logmanager"
	"github.com/ligato/cn-infra/statuscheck"
)

// Flavour glues together multiple plugins that are useful for almost every micro-service
type Flavour struct {
	Logrus       logrus.Plugin
	HTTP         httpmux.Plugin
	LogManager   logmanager.Plugin
	ServiceLabel servicelabel.Plugin
	StatusCheck  statuscheck.Plugin

	injected bool
}

// Inject sets object references
func (f *Flavour) Inject() error {
	if f.injected {
		return nil
	}

	f.HTTP.LogFactory = &f.Logrus
	f.LogManager.ManagedLoggers = &f.Logrus
	f.LogManager.HTTP = &f.HTTP
	f.StatusCheck.HTTP = &f.HTTP

	f.injected = true

	return nil
}

// Plugins combines all Plugins in flavour to the list
func (f *Flavour) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
