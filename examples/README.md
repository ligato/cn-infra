# CN-infra examples

The examples folder contains several executable examples (built from their 
respective `main.go` files) used to illustrate the cn-infra functionality. 
While most of the examples show a very simple use case, they still often
need to connect to ETCD/Redis and/or Kafka. Therefore, you need to have
instances of Etcd and Kafka running prior to starting examples.

Examples with the suffix `_lib` demonstrate the usage of CN-Infra APIs in 
generic Go programs. You can simply import the CN-Infra library where the
API is declared into your program and start using the API.

Examples with the suffix `_plugin` demonstrate the usage of CN-Infra APIs
within the context of plugins. Plugins are the basic building blocks
of any given CN-Infra application.  The CNB-Infra plugin framework
provides plugin initialization and graceful shutdown and supports
uniform dependency injection mechanism to manage dependencies between
plugins.

Current examples:
* **[cassandra lib](cassandra_lib)** shows how to use the Cassandra data 
  broker API to access the Cassandra database,
* **[datasync plugin](datasync_plugin)** demonstrates the usage
  of the data synchronization APIs of the datasync package inside
  an example plugin,
* **[etcdv3 lib](etcdv3_lib)** shows how to use the ETCD data broker API 
  to write data into ETCD and catch this change as an event by the watcher,
* **[flags plugin](flags_plugin/main.go)** registers flags and shows their 
  runtime values in an example plugin,
* **[kafka lib](kafka_lib)** shows how to use the Kafka messaging library
  on a set of individual tools (sync and async producer, consumer, mux),
* **[kafka plugin (non-clustered)](kafka_plugin/non_clustered/main.go)** 
  contains a simple plugin which registers a Kafka consumer and sends
  a test notification,
* **[kafka plugin (clustered)](kafka_plugin/clustered/main.go)** contains 
  a simple plugin which registers a Kafka consumer watching on specific
  partition/offset and sends a test notification to partition in Kafka
  cluster,
* **[logs lib](logs_lib)** shows how to use the logger library and switch 
  between the log levels,
* **[logs plugin](logs_plugin)** shows how to use the logger library
  in a simple plugin,
* **[redis lib](redis_lib)** contains several examples that use
  the Redis data broker API,
* **[model](model)** shows how to define a custom data model using
  Protocol Buffers and how to integrate it into an application,
* **[simple-agent](simple-agent)** demonstrates how easily a set of
  CN-infra based plugins can be turned into an application.