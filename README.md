# CN-Infra

[![Build Status](https://travis-ci.org/ligato/cn-infra.svg?branch=master)](https://travis-ci.org/ligato/cn-infra)
[![Coverage Status](https://coveralls.io/repos/github/ligato/cn-infra/badge.svg?branch=master)](https://coveralls.io/github/ligato/cn-infra?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/ligato/cn-infra)](https://goreportcard.com/report/github.com/ligato/cn-infra)
[![GoDoc](https://godoc.org/github.com/ligato/cn-infra?status.svg)](https://godoc.org/github.com/ligato/cn-infra)
[![GitHub license](https://img.shields.io/badge/license-Apache%20license%202.0-blue.svg)](https://github.com/ligato/cn-infra/blob/master/LICENSE.md)

CN-Infra (cloud-native infrastructure) is a Golang platform for building
custom management/control plane applications for cloud-native Virtual 
Network Functions (VNFs). Cloud-native VNFs are also known as "CNFs". 

## Cloud-Native Virtual Network Functions (CNFs)
So what is a *cloud-native* virtual network function? 

A virtual network function (or VNF), as commonly known today, is a software
implementation of a network function that runs on one or more *virtual 
machines* (VMs) on top of the hardware networking infrastructure — routers,
switches, etc. Individual virtual network functions can be connected or
combined together as building blocks to offer a full-scale networking 
communication service. A VNF may be implemented as standalone entity using
existing networking and orchestration paradigms - for example being 
managed through CLI, SNMP or Netconf. Alternatively, an NFV may be a part
of an SDN architecture, where the control plane resides in an SDN 
controller and the data plane is implemented in the VNF.

A *cloud-native VNF* is a VNF designed for the emerging cloud environment -
it runs in a container rather than a VM, its lifecycle is orchestrated 
by a container orchestration system, such as Kubernetes, and it's using
cloud-native orchestration paradigms. In other words, its control/management
plane looks just like any other container based [12-factor app][1]. to 
orchestrator or external clients it exposes REST or gRPC APIs, data stored
in centralized KV data stores, communicate over message bus, cloud-friendly
logging and config, cloud friendly build & deployment process, etc.,
Depending on the desired functionality, scale and performance, a cloud-
native VNF may provide a high-performance data plane, such as the [VPP][2].


## The CN-Infra Platform Architecture

Each management/control plane app built on top of the CN-Infra platform is 
basically a set of modules called "plugins" in CN-Infra lingo, where each 
plugin provides a very specific/focused functionality. Some plugins are 
provided by the CN-Infra platform itself, some are written by the app's 
implementors. In other words, the CN-Infra platform itself is implemented
as a set of plugins that together provide the platform's functionality, 
such as logging, health checks, messaging (e.g. Kafka), a common front-end
API and back-end connectivity to various KV data stores (Etcd, Cassandra, 
Redis, ...), and REST and gRPC APIs. 

The architecture of the CN-Infra platform is shown in the following figure.

![arch](docs/imgs/high_level_arch_cninfra.png "High Level Architecture of cn-infra")

The CN-Infra platform consists of a **[Core](core)** that provides plugin
lifecycle management (initialization and graceful shutdown of plugins) 
and a set of platform plugins. Note that the figure shows not only 
CN-Infra plugins that are a part of the CN-Infra platform, but also 
app plugins that use the platform. CN-Infra platform plugins provide 
APIs that are consumed by app plugins. App plugins themselves may 
provide their own APIs consumed by external clients.

The platform is modular and extensible. Plugins supporting new functionality
(e.g. another KV store or another message bus) can be easily added to the
existing set of CN-Infra platform plugins. Moreover, CN-Infra based apps
can be built in layers: a set of app plugins together with CN-Infra plugins
can form a new platform providing APIs/services to higher layer apps. 
This approach was used in the [VPP Agent][3] - a management/control agent
for [VPP][2] based software data planes.,

Extending the code base does not mean that all plugins end up in all 
apps - app writers can pick and choose only those platform plugins that 
are required by their app; for example, if an app does not need a KV 
store, the CN-Infra platform KV data store plugins would not be included
in the app. All plugins used in an app are statically linked into the 
app.

## CN-Infra Plugins
A CN-Infra plugin is typically implemented as a library providing the 
plugin's functionality/APIs wrapped in a plugin wrapper. A CN-Infra 
library can also be used standalone in 3ed party apps that do not use
the CN-Infra platform. The plugin wrapper provides lifecycle management 
for the plugin component.

Platform plugins in the current CN-Infra release provide functionality
in one of the following functional areas:

* **RPC** - allows to expose application's API via REST or gRPC:
    * [HTTPmux](httpmux) -  HTTP requests and allows app plugins to define
      their own REST APIs.
        
* **Data Stores** - provides a common data store API for app plugins (the 
    Data Broker) and back-end clients for Etcd, Redis and Cassandra. The 
    data store related plugin are as follows:
  - [Etcd](db/keyval/etcdv3) - implements keyval skeleton provides access 
    to etcd
  - [Redis](db/keyval/redis) - implements keyval skeleton provides access
    to redis
  - [Casssandra](db/sql/cassandra) -
    
* **Messaging** - provides a common API and connectivity to message buses:
    - Kafka](messaging/kafka) - provides access to Kafka brokers
    
* **Logging**:
    * [Logrus wrapper](logging/logrus) - implements logging skeleton 
      using the Logrus library. An app writer can create multiple loggers -
      for example, each app plugin can have its own logger. Log level
      for each logger can be controlled individually at run time through
      the Log Manager REST API.
    * [Log Manager](logging/logmanager) - allows the operator to set log
      level for each logger using a REST API.
    
* **[Health](statuscheck)** - Self health check mechanism between plugins 
    plus RPCs:
    - [StatusCheck](statuscheck) - allows to monitor the status of plugins
      and exposes it via HTTP
    - Probes (callable remotely from K8s)
  
* **Miscellaneous** - value-add plugins supporting the operation of a 
    CN-Infra based application: 
  - [Datasync](datasync/resync) - provides data resynchronization after HA 
    events (restart or connectivity restoration after an outage) for data
    stores, gRPC and REST.
  - [ServiceLabel](servicelabel) - provides setting and retrieval of a 
    unique identifier for a CN-Infra based app. A cloud app typically needs
    a unique identifier so that it can differentiated from other instances 
    of the same app or from other apps (e.g. to have its own space in a kv 
    data store).
   
## Quickstart
The following code shows the initialization/start of a simple agent 
application built on the CN-Infra platform. The code for this example
can be found [here](examples/simple-agent/agent.go).
```
func main() {
	flavour := Flavour{}
	agent := core.NewAgent(logroot.Logger(), 15*time.Second, flavour.Plugins()...)

	err := core.EventLoopWithInterrupt(agent, nil)
	if err != nil {
		os.Exit(1)
	}
}
```

## Documentation

GoDoc can be browsed [online](https://godoc.org/github.com/ligato/cn-infra).

## Contributing

If you are interested in contributing, please see the [contribution guidelines](CONTRIBUTING.md).

[1]: https://12factor.net/
[2]: https//fd.io
[3]: https://github.com/ligato/vpp-agent