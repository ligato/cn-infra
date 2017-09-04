# CN-infra examples

There are several `main.go` files used as an illustration of the cn-infra
functionality. Most of the examples show a very simple use case using the
real ETCD and/or Kafka plugins, so for specific examples, they have to be
started first.

Current examples:
* **[cassandra lib](cassandra_lib)** shows how to use the Cassandra data broker API
  to access the Cassandra database,
* **[datasync plugin](datasync_plugin)** showcases the data synchronization APIs of
  the datasync package on an example plugin,
* **[etcdv3 lib](etcdv3_lib)** shows how to use the ETCD data broker API 
  to write data into ETCD and catch this change as an event by the watcher,
* **[flags plugin](flags_plugin/main.go)** registers flags and shows their 
  runtime values in an example plugin,
* **[kafka lib](kafka_lib)** shows how to use the Kafka messaging library
  on set of individual tools (producer, sync and async consumer),
* **[kafka plugin](kafka_plugin/main.go)** contains a simple plugin which registers a 
  Kafka consumer and sends a test notification,
* **[logs lib](logs_lib)** shows how to use the logger library and switch between 
  the log levels,
* **[logs plugin](logs_plugin)** showcases how to ue the logger library in a 
  simple plugin,
* **[redis lib](redis_lib)** contains several examples that use the Redis data 
  broker API,
* **[model](model)** show how to define a custom data model using Protocol Buffers
  and how to integrate it into the application,
* **[simple-agent](simple-agent)** showcases an approach how a set of plugins can 
  be turned to an application.

## How to run an example

 **1. Start ETCD server on localhost**

  ```
  sudo docker run -p 2379:2379 --name etcd --rm \
  quay.io/coreos/etcd:v3.0.16 /usr/local/bin/etcd \
  -advertise-client-urls http://0.0.0.0:2379 \
  -listen-client-urls http://0.0.0.0:2379
  ```

 **2. Start Kafka**

 ```
 sudo docker run -p 2181:2181 -p 9092:9092 --name kafka --rm \
  --env ADVERTISED_HOST=172.17.0.1 --env ADVERTISED_PORT=9092 spotify/kafka
 ```

 **3. Start the desired example**

 Each example can be started now from its directory.
 ```
 go run main.go  \
 --etcdv3-config=/opt/vnf-agent/dev/etcd.conf \
 --kafka-config=/opt/vnf-agent/dev/kafka.conf
 ```
