# Cassandra implementation of Broker interface

The API was tested with Cassandra 3 and supports:
- UDT (User Defined Types) / embedded structs and honors gocql.Marshaler/gocql.Unmarshaler
- handling all primitive types (like int aliases , IP address); 
  - net.IP can be stored as ipnet
  - net.IPNet can be stored with a MarshalCQL/UnmarshalCQL wrapper go structure
- dumping all rows except for Table
- quering by secondary indexes
- mocking of gocql behavior (using gockle library)in automated unit tests

# Cassandra Timeouts

The API will allow the client to configure either single node or multi-node cluster.
Also, the client can configure following timeouts:
- ConnectTimeout
    - Initial connection timeout, used during initial dial to server
    - Default value is 600ms
- Timeout (Query timeout)
    - Connection timeout, used during executing query
    - Default value is 600ms
- ReconnectInterval
    - If not zero, gocql attempt to reconnect known DOWN nodes in every ReconnectSleep (ReconnectInterval)
    - Default value is 60s

# Cassandra Data Consistency

The API will allow the client to configure consistency level for both
- Session
- Query (to be implemented)

Factors to consider for achieving desired consistency level.
- Replication strategy
    - A replication strategy determines the nodes where replicas are placed.
        - SimpleStrategy: Use for a single data center only.
        - NetworkTopologyStrategy: Used for more than one data center.
- Replication factor
    - The total number of replicas across the cluster is referred to as the replication factor.
    - A replication factor of 1 means that there is only one copy of each row on one node.
    - A replication factor of 2 means that there are two copies of each row, and each copy is stored on a different node.
    - The replication factor should not exceed the number of nodes in the cluster.

- To achieve Quorum
    - Quorum = (sum_of_replication_factors / 2) + 1
    - (nodes_written + nodes_read) > replication_factor

