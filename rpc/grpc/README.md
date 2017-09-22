# GRPC Plugin

The `GRPC Plugin` is a infrastructure Plugin which allows app plugins 
to handle GRPC requests (see the diagram below) in this sequence:
1. GRPC Plugin starts the GRPC server + net listener in its own goroutine
2. Plugins register their handlers with `GRPC Plugin`.
   To service GRPC requests, a plugin must first implement a handler
   function and register it at a given URL path using
   the `RegisterService` method. `GRPC Plugin` uses an GRPC request
   multiplexer from the `grpc/Server`.
3. GRPC server routes GRPC requests to their respective registered handlers
   using the `grpc/Server`.

![http](../../docs/imgs/grpc.png)

**Configuration**

- the server's port can be defined using commandline flag `grpc-port` or 
  via the environment variable GRPC_PORT.

**Example**

The following example demonstrates the usage of the `GRPC Plugin` plugin
API:
```
// Register our GRPC request handler:
GRPC.RegisterService(descriptor, service)
```

TODO you can expose GRPC services also through HTTP if you optionally inject GRPC.Dep.HTTP

Once the handler is registered with `GRPC Plugin` and the agent is running, 
you can use `curl` to verify that it is operating properly:
```
$ curl -X GET http://localhost:9191/service/example
{
  "Example": "This is an example"
}
```
