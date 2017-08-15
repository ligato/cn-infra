# Flavors

A flavor is a reusable collection of plugins with initialized [dependencies](../docs/guidelines/PLUGIN_DEPENDENCIES.md). 
CN-Infra provides the following [flavors](../docs/guidelines/PLUGIN_FLAVORS.md):
* [generic flavor](generic) - a collection of plugins that are useful for almost
  every micro-service
* [etcd + Kafka flavor](etcdkafka) - adds etcd & Kafka client plugin instances to 
  the generic flavor 
  
The following diagram shows:
* plugins that are part of the flavor
* initialized (injected) [statuscheck](../statuscheck) dependency 
  inside [etcd client plugin](../db/keyval/etcdv3) and [Kafka client plugin](../messaging/kafka)
* [etcd + Kafka flavor](etcdkafka) extends [generic flavor](generic) 

![flavors](../docs/imgs/flavors.png)
