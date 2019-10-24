# grpc-plugin

Start the `grpc-server` example typing following in the grpc-server folder:
```
go run main.go 
```

In order to pass custom configuration file (for example the `grpc.conf` located in the same directory), start grpc-server with following parameter:
```
go run main --grpc-config=<config-filepath>
```
Note: `main.go` must be edited in this case because it makes use of the direct config injection via `grpc.UseConf` method, which overrides any provided config file. To enable it, remove method mentioned. 

Start the `grpc-client` exampl typing following in the grpc-client folder:
```
go run main.go
```

