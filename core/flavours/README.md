# Flavour

Flavour in this context represents a collection of plugins. The selection of plugins defines capabilities
available for an agent. The collection that is available out-of-box is called generic flavour.

See also package core [readme](../README.md).

## Generic flavour

The generic flavour groups commonly used plugins. It currently contains the following plugins:

- log plugin
- etcd plugin
- kafka plugin

Usage of the generic flavour consists of the following steps:

1. Create new instance of the flavour

    ```f := generic.Flavour{}```
  
2. Register flags that are supposed to be parsed from commandline options (e.g: config files).
   
    ```f.RegisterFlags()```
  
3. Create instances of plugins using the parsed config.
   
    ```f.ApplyConfig()```

4. Interconnect the plugins and inject the dependencies.
   
   ```f.Inject()```

5. Retrieve the plugins from the flavour and pass them to the agent constructor.

   ```
       pl := f.Plugins
       agent := core.NewAgent(logger, timeout, pl...)
   ```
    
**Note:**
Steps 2 and 3 can be skipped in tests and plugin can be instantiated using a different mechanism.
 
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
