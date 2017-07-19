# Flavour

Flavour in this context represents a collection of plugins. The selection of plugins defines capabilities
available for an agent. The collection that is available out-of-box is called generic flavour.

See also package core [readme](../README.md).

## Generic flavour

The generic flavour groups commonly used plugins. It currently contains the following plugins:

- log plugin
- etcd plugin
- kafka plugin

 
The flavour can be reused in order to build more complex agent.

### Command line arguments

The generic flavour reads the following commandline/env variable arguments:
- `etcdv3-config` Location of the Etcd configuration file; also set via 'ETCDV3_CONFIG' env variable.
- `kafka-config` Location of the Kafka configuration file; also set via 'KAFKA_CONFIG' env variable.

### Extending the generic flavour

The following interfaces are exposed by plugins included in the generic flavour:

- LogPlugin

     `logging.LogFactory`
     `logging.LogManagement`

- EtcdPlugin

     `keyval.KvPlugin`

- KafkaPlugin

     `kafka.Mux`
