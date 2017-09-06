# Flavors

A [flavors](../docs/guidelines/PLUGIN_FLAVORS.md) is a reusable collection of plugins 
with initialized [dependencies](../docs/guidelines/PLUGIN_DEPENDENCIES.md). 

Most importatnt CN-Infra flavors:
* [local flavor](local) - a minimal set of plugins. It just initializes logging & statuchek.
  It is useful for embedding agent plugins to different projects that use their own infrasturcure.
* [RPC flavor](rpc) - a collection of plugins that exposes RPCs. It also register management API for:
  * status check (RPCs probed from systems such as K8s)
  * logging (for changing log level at runtime remotely)
* [all connectors flavor](connectors/all_connectors_flavor.go) - is combination of ETCD, Cassandra, Redis & Kafka related plugins.
  
The following diagram depicts:
* plugins that are part of a specific flavor
* initialized (injected) [statuscheck](../health/statuscheck) dependency 
  inside [etcd client plugin](../db/keyval/etcdv3) and [Kafka client plugin](../messaging/kafka)
* embedded [local flavor](local) in:
    * [all connectors flavor](connectors) 
    * [RPC flavor](rpc)

![flavors](../docs/imgs/flavors.png)
