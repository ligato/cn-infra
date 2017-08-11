# Flavors

A flavor is a reusable collection of plugins with initialized [dependencies](../docs/guidelines/PLUGIN_DEPENDENCIES.md). 
CN-Infra provides the following [flavors](../docs/guidelines/PLUGIN_FLAVORS.md):
* [generic flavor](generic) - a collection of plugins that are useful for almost
  every micro-service
* [etcd + kafka flavor](etcdkafka) - adds etcd & kafka client plugin instances to 
  the generic flavor 
  
In following diagram ilustrates:
* plugins that are part of the flavor
* initialized (injected) [statuscheck](../statuscheck) dependency 
  inside [etcd client plugin](../db/keyval/etcdv3) and [kafka client plugin](../messaging/kafka)
* [etcd + kafka flavor](etcdkafka) extends [generic flavor](generic) 

![flavors](../docs/imgs/flavors.png)