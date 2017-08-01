# Concept
It is usual that logic of a microservice relies on data synchronization. In this context 
the data synchronization is about two or more data sets that needs to be synchronized when 
event was published. Event can be published by database (when particular data was changes),   
by message bus (like Kafka).

The datasync plugin helps other plugins/APP to (see next diagram):
 1. Watch data - subscribe for particular data changes to receive events 
                 with current data in channels
 2. Publish data - publish events with current values
 
![datasync](../docs/imgs/datasync.png)
 
## Watch data
The API distinguishes between:
1. Full data resync (startup, for certain fault recovery) - reprocess whole (configuration) data 
2. Optimized datasync - process only changes (deltas)

### Full data resync
Fault tolerance vs. data synchronization
In a fault-tolerant solution, there is a need to recover from faults. This plugin helps to solve the
data resynchronization (data RESYNC) aspect of fault recovery.
When the Agent looses connectivity (to ETCD, Kafka, VPP etc.), it needs to recover from that fault:
1. When connectivity is reestablished after a failure, the agent can resynchronize (configuration) from a northbound 
   client with southbound configuration (VPP etc.).
2. Alternative behavior: Sometimes it is easier to use "brute force" and restart the container (both VPP and the agent) 
   and skip the resynchronization. This restart is supposed to be done by control plane & orchestration
   layer. The Agent is supposed to just publish the event.

To report a fault/error occurred and notify the datasync plugin there is defined following API call.
```
TODO example how to process the data from channel
```

### Optimized mode
In optimized mode we do not need to reprocess whole (configuration) data but rather process just the delta
(only the changed object in current value of go channel event).
```
TODO example how to process the data from channel
```
 
## Responsibility of plugins
Each plugin is responsible for its own part of (configuration) data received from northbound clients. Each plugin needs 
to be decoupled from a particular datasync transport (ETCD/Redis, more will come soon: Kafka, Local, GRPC, REST ...).
Every other plugin (then datasync plugin) receives (configuration) data only through GO interfaces defined 
in [datasync_api.go](datasync_api.go)

The data of one plugin can have references to data of another plugin. Therefore, we need 
to have proper time/order of data resynchronization between plugins. The datasync plugin
initiates full data resync in the same order as the other plugins have been registered in Init().

```
TODO registration in Init() example
```