# Datasync example

To start the examples you have to have ETCD running first.
if you don't have it installed locally you can use the following docker
image.
```
sudo docker run -p 2379:2379 --name etcd --rm \
    quay.io/coreos/etcd:v3.0.16 /usr/local/bin/etcd \
    -advertise-client-urls http://0.0.0.0:2379 \
    -listen-client-urls http://0.0.0.0:2379
```

It will bring up ETCD listening on port 2379 for client communication.

In the example, the location of the ETCD configuration file is defined
with the `-etcdv3-config` argument or through the `ETCDV3_CONFIG`
environment variable.
By default, the application tries to connect to ETCD running
on the localhost and port 2379.

To run the example, type:
```
go run main.go deps.go [-etcdv3-config <config-filepath>]
```

