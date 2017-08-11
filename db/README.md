# Concept
CN-infra DB concept (see following diagram) is based on: 
* Broker API - used by app plugins to PULL (meaning running queries), PUSH (meaning executing write operations)
* Watcher API - used by app plugins to WATCH (meaning monitor changes and be notified as soon as the change occurred)

The Broker & Watcher APIs abstract common operations among different databases (ETCD, Redis, Cassandra).
Still, there are major differences between [keyval](keyval)-base & [sql](sql)-based databases.
Therefore there are defined Broker & Watcher GO interfaces in both packages while using same method names
and similar arguments.

![db](../docs/imgs/db.png)