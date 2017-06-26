package keyval

import "github.com/ligato/cn-infra/core"

// KvPlugin provides unifying interface for different key-value datastore implementations.
type KvPlugin interface {
	core.Plugin
	// NewRootBroker returns a ProtoBroker providing methods for retrieving and editing key-value pairs in datastore.
	NewRootBroker() ProtoBroker
	// NewRootWatcher returns a ProtoWatcher providing for watching the changes in datastore.
	NewRootWatcher() ProtoWatcher
	// NewPrefixedBroker returns a ProtoBroker instance that prepends given keyPrefix to all keys in its calls.
	NewPrefixedBroker(keyPrefix string) ProtoBroker
	// NewPrefixedWatcher returns a ProtoWatcher instance. Given key prefix is prepended to keys during watch subscribe phase.
	// The prefix is removed from the key retrieved by GetKey() in ProtoWatchResp.
	NewPrefixedWatcher(keyPrefix string) ProtoWatcher
}

type kvPlugin struct {
	KvPlugin
}

// NewKvPlugin creates new instance of kv plugin with given key-val data store connection.
func NewKvPlugin(conn KvPlugin) KvPlugin {
	return &kvPlugin{conn}
}

func (k *kvPlugin) Init() error {
	return k.KvPlugin.Init()
}

func (k *kvPlugin) Close() error {
	return k.KvPlugin.Close()
}
