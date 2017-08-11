# Concept
Package datasync defines the interfaces for the abstraction [datasync_api.go](datasync_api.go)
of  a data synchronization between app plugins and different backend data sources 
(such/as data stores, message buses, or RPC-connected clients).

In this context, the data synchronization is about multiple data sets 
that need to be synchronized whenever a particular event is published. 
The event can be published by:
- database (when particular data was changed); 
- by message bus (such as consuming messages from Kafka topics); 
- or by RPC clients (when GRPC or REST service call ).

The data synchronization APIs are centered around watching 
and publishing data change events. These events are processed asynchronously.

The data handled by one plugin can have references to the data of another plugin. 
Therefore, proper time/order of data resynchronization between plugins. The datasync plugin
initiates full data resync in the same order as the other plugins have been registered in Init().
  
## Watch data API
Watch data API is used by app plugin (see following diagram) to:
1. Subscribe channels for particular data changes `WatchData()` 
while being abstracted from a particular message source (data store, message bus or RPC)
2. Process Full data RESYC (startup, for certain fault recovery) event reprocess whole data set.
3. Feedback is given to the user of this API (e.g. successful configuration or an error) by callback.
4. Process Incremental data CHANGE. It is Optimized mode that 
   works only with minimal set of changes (deltas).
5. Feedback is again given to the user of this API (e.g. successful configuration or an error) by callback.

![datasync](../docs/imgs/datasync_watch.png)

This APIs defines two types of events that a plugin must be able to process:
1. Full Data RESYNC (resynchronization) event is defined to trigger
   resynchronization of the whole configuration. This event is used
   after agent start/restart, or when agent's connectivity to an
 external data source/sink is lost and restored.
2. Incremental Data CHANGRE event is defined to trigger incremental processing of
   configuration changes. Data changes events are sent after the data
   resync has completed. Each data change event contains both the
   previous and the new/current values for the data. The Data synchronization 
   is switched to optimized mode after successful FULL data RESYNC. 


Full data RESYNC is needed during startup or fault recovery. 

See the [example](examples/simple_watch)

### Full data RESYC
In a fault-tolerant solution, there is a need to recover from faults. This plugin helps to solve the
data resynchronization (data RESYNC) aspect of fault recovery.
When the Agent loses connectivity (to etcd, Kafka, VPP etc.), it needs to recover from that fault:
1. When connectivity is reestablished after a failure, the agent can resynchronize (configuration) from a northbound 
   client with southbound configuration (VPP etc.).
2. Alternative behavior: Sometimes it is easier to use "brute force" and restart the container (both VPP and the agent) 
   and skip the resynchronization. This restart is supposed to be done by control plane & orchestration
   layer. The Agent is supposed to just publish the event.

TODO To report a fault/error occurred and notify the datasync plugin there is defined following API call.

### Incremental data CHANGE
In this optimized mode, we do not need to reprocess the whole (configuration) data but rather process just the delta
(only the changed object in current value of go channel event).
 
## Publish data API

Publish data API is used by app plugins to asynchronously publish events 
with particular data changes values and still abstract from data store, message bus, local/RPC client.

![datasync](../docs/imgs/datasync_publish.png)
