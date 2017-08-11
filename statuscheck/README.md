# Concept

The `statuscheck` infrastructure plugin monitors overall status of the agent 
and form partial statuses of agents plugins.
The status is exposed to external clients via [ETCD - datasync](../datasync) 
and [HTTP](../httpmux) (see following diagram).

![status check](../docs/imgs/status_check.png)

For more detailed description see the godoc (especially [doc.go](doc.go)).

**Overall Agent Status**

Overall Agent Status is aggregated from all Plugins Status (logical AND foreach Plugin Status success/error).

To retrieve the current overall status of the agent from ETCD, get the 
`/vnf-agent/<agent-label>/check/status` key:
```
$ etcdctl get /vnf-agent/<agent-label>/check/status/v1/agent
/vnf-agent/<agent-label>/check/status/v1/agent
{"build_version":"e059fdfcd96565eb976a947b59ce56cfb7b1e8a0","build_date":"2017-06-16.14:59","state":1,"start_time":1497617981,"last_change":1497617981,"last_update":1497617991}
```

To verify the agent status via HTTP (e.g. for Kubernetes 
[liveness and readiness probes][1], use the `/liveness` and `/readiness` URLs:
```
$ curl -X GET http://localhost:9191/liveness
{"build_version":"e059fdfcd96565eb976a947b59ce56cfb7b1e8a0","build_date":"2017-06-16.14:59","state":1,"start_time":1497617981,"last_change":1497617981,"last_update":1497617991}
$ curl -X GET http://localhost:9191/readiness
{"build_version":"e059fdfcd96565eb976a947b59ce56cfb7b1e8a0","build_date":"2017-06-16.14:59","state":1,"start_time":1497617981,"last_change":1497617981,"last_update":1497617991}
```

To change the HTTP server port (default `9191`), us the `http-port` 
option of the agent, e.g.:
```
$ vpp-agent -http-port 9090
```

**Plugins Status**

Plugin Status can be PUSHed by a multiple plugins but there  
is posible PULL based approach when status check plugin
periodically probes previously registered plugins.

To retrieve the current status of a plugin from ETCD, get the 
`/vnf-agent/<agent-label>/check/status/v1/plugin/<PLUGIN_NAME>` key, e.g. for GoVPP:

```
$ etcdctl get /vnf-agent/<agent-label>/check/status/v1/plugin/GOVPP
/vnf-agent/<agent-label>/check/status/v1/plugin/GOVPP
{"state":2,"last_change":1496322205,"last_update":1496322361,"error":"VPP disconnected"}
```

PUSH Plugin Status:
![status check pull](../docs/imgs/status_check_pull.png)

PULL Plugins Status - PROBING:
![status check push](../docs/imgs/status_check_push.png)


[1]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/