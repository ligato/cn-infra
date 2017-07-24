# Examples for cn-infra/db/keyval/redis
* simple
  * Straight API calls, no particular scenario
  * Also includes function generateSampleConfigs() to show how to generate each type of client configurations. 
* airport
  * Simulates aiport operations with four functional modules - arrival, departure, runway and hangar
  * Communicates and updates flight info through Redis.
  * Displays flight status, live.

To run these examples, you must have access to a Redis installation, either locally or remotely.

You can download docker images from:
* [https://github.com/AliyunContainerService/redis-cluster](https://github.com/AliyunContainerService/redis-cluster)
  * Redis cluster with Sentinel
* [https://github.com/Grokzen/docker-redis-cluster](https://github.com/Grokzen/docker-redis-cluster)
  * Redis standalone and cluster

See sample configurations to use with the API here:
* [https://github.com/ligato/cn-infra/db/keyval/redis/examples/node-client.yaml](https://github.com/ligato/cn-infra/db/keyval/redis/examples/node-client.yaml)
* [https://github.com/ligato/cn-infra/db/keyval/redis/examples/cluster-client.yaml](https://github.com/ligato/cn-infra/db/keyval/redis/examples/cluster-client.yaml)
* [https://github.com/ligato/cn-infra/db/keyval/redis/examples/sentinel-client.yaml](https://github.com/ligato/cn-infra/db/keyval/redis/examples/sentinel-client.yaml)

Modify the configuration to fit your environment.

To start examples, 'cd' to cn-infra/db/keyval/redis/examples directory.  Then execute
```
go run simple/simple.go -n node-client.yaml
```
```
go run simple/simple.go -c cluster-client.yaml
```
```
go run simple/simple.go -s sentinel-client.yaml
```
```
go run airport/airport.go -n node-client.yaml
```
```
go run airport/airport.go -c cluster-client.yaml
```
```
go run airport/airport.go -s sentinel-client.yaml
```
