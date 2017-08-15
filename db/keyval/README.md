# Key-value datastore

The `keyval` package defines the API for access to key-value data store. 
The `Broker` interface supports reading and manipulation of key-value 
pairs. The `Watcher` API provides functions for monitoring of changes in 
a data store. Both interfaces are available with arguments of type 
`[]bytes` and `proto.Message`.

The package also provides a skeleton for a key-value plugin. The particular 
data store is selected in the `NewSkeleton` constructor using an argument
of type `CoreBrokerWatcher`. The skeleton handles the plugin's life-cycle
and provides unified access to datastore implementing the `KvPlugin` 
interface.
