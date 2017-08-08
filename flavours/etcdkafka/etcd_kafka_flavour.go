package etcdkafka

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/flavours/generic"
	"github.com/ligato/cn-infra/messaging/kafka"
)

// Flavour glues together generic.Flavour plugins with:
// - ETCD (useful for watching config.)
// - Kafka plugins (useful for publishing events)
type Flavour struct {
	Generic generic.Flavour
	Etcd    etcdv3.Plugin
	Kafka   kafka.Plugin

	injected bool
}

// Inject sets object references
func (f *Flavour) Inject() error {
	if f.injected {
		return nil
	}

	f.Generic.Inject()

	f.Etcd.LogFactory = &f.Generic.Logrus
	f.Etcd.ServiceLabel = &f.Generic.ServiceLabel
	f.Kafka.LogFactory = &f.Generic.Logrus
	f.Kafka.ServiceLabel = &f.Generic.ServiceLabel

	f.injected = true

	return nil
}

// Plugins combines all Plugins in flavour to the list
func (f *Flavour) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
