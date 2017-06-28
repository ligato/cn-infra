# Key-value datastore

The package defines API for access to key-value data store. `Broker` interface allows to read and manipulate key-value pairs.
`Watcher` provides functions for monitoring of changes in a datastore. Both interfaces are available with arguments
 of type `[]bytes` and `proto.Message`.