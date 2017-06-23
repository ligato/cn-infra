package keyval

import "github.com/ligato/cn-infra/core"

type Plugin interface {
	core.Plugin
	NewRootBroker() ProtoBroker
	NewRootWatcher() ProtoWatcher
	NewPrefixedBroker(keyPrefix string) ProtoBroker
	NewPrefixedWatcher(keyPrefix string) ProtoWatcher
}
