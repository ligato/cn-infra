package cassandra

import (
	"github.com/gocql/gocql"
	"strings"
	"time"
)

// ClientConfig Configuration for Cassandra clients
type ClientConfig struct {

	// A list of host addresses of cluster nodes.
	Hosts []string `json:"hosts"`

	// port for Cassandra (default: 9042)
	Port int `json:"port"`

	// connection timeout (default: 600ms)
	QueryTimeout time.Duration `json:"query_timeout"`

	// initial connection timeout, used during initial dial to server (default: 600ms)
	InitialConnectTimeout time.Duration `json:"initial_connect_timeout"`

	// If not zero, gocql attempt to reconnect known DOWN nodes in every ReconnectSleep.
	NodeDownReconnectInterval time.Duration `json:"node_down_reconnect_interval"`

	// ProtoVersion sets the version of the native protocol to use, this will
	// enable features in the driver for specific protocol versions, generally this
	// should be set to a known version (2,3,4) for the cluster being connected to.
	//
	// If it is 0 or unset (the default) then the driver will attempt to discover the
	// highest supported protocol for the cluster. In clusters with nodes of different
	// versions the protocol selected is not defined (ie, it can be any of the supported in the cluster)
	ProtoVersion int `json:"proto_version"`

	//Initial Keyspace
	Keyspace string `json:"keyspace"`
}

// CreateSessionFromClientConfigAndKeyspace Creates session from given configuration and keyspace
func CreateSessionFromClientConfigAndKeyspace(config ClientConfig, setKeyspace bool) (*gocql.Session, error) {

	gocqlClusterConfig := gocql.NewCluster(HostsAsString(config.Hosts))
	gocqlClusterConfig.Port = config.Port
	gocqlClusterConfig.ConnectTimeout = config.InitialConnectTimeout * time.Millisecond
	gocqlClusterConfig.Timeout = config.QueryTimeout * time.Millisecond
	gocqlClusterConfig.ReconnectInterval = config.NodeDownReconnectInterval * time.Second
	gocqlClusterConfig.ProtoVersion = config.ProtoVersion

	if setKeyspace {
		gocqlClusterConfig.Keyspace = config.Keyspace
	}

	session, err := gocqlClusterConfig.CreateSession()

	if err != nil {
		return nil, err
	}

	return session, nil
}

// HostsAsString converts an array of hosts addresses into a comma separated string
func HostsAsString(hostArr []string) string {
	return strings.Join(hostArr, ",")
}
