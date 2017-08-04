package cassandra

import (
	"strings"
	"time"

	"github.com/gocql/gocql"
)

// ClientConfig Configuration for Cassandra clients loaded from a configuration file
type Config struct {

	// A list of host addresses of cluster nodes.
	Hosts []string `json:"hosts"`

	// port for Cassandra (default: 9042)
	Port int `json:"port"`

	// connection timeout (default: 600ms)
	Timeout time.Duration `json:"timeout"`

	// initial connection timeout, used during initial dial to server (default: 600ms)
	ConnectTimeout time.Duration `json:"connect_timeout"`

	// If not zero, gocql attempt to reconnect known DOWN nodes in every ReconnectSleep.
	ReconnectInterval time.Duration `json:"reconnect_interval"`

	// ProtoVersion sets the version of the native protocol to use, this will
	// enable features in the driver for specific protocol versions, generally this
	// should be set to a known version (2,3,4) for the cluster being connected to.
	//
	// If it is 0 or unset (the default) then the driver will attempt to discover the
	// highest supported protocol for the cluster. In clusters with nodes of different
	// versions the protocol selected is not defined (ie, it can be any of the supported in the cluster)
	ProtoVersion int `json:"proto_version"`
}

// ClientConfig wrapping gocql ClusterConfig
type ClientConfig struct {
	*gocql.ClusterConfig
}

const defaultTimeout = 600 * time.Millisecond
const defaultConnectTimeout = 600 * time.Millisecond
const defaultReocnnectInterval = 60 * time.Second
const defaultProtoVersion = 4

// ConfigToClientConfig transforms the yaml configuration into ClientConfig.
func ConfigToClientConfig(ymlConfig *Config) (*ClientConfig, error) {

	timeout := defaultTimeout
	if ymlConfig.Timeout > 0 {
		timeout = ymlConfig.Timeout
	}

	connectTimeout := defaultConnectTimeout
	if ymlConfig.ConnectTimeout > 0 {
		connectTimeout = ymlConfig.ConnectTimeout
	}

	reconnectInterval := defaultReocnnectInterval
	if ymlConfig.ReconnectInterval > 0 {
		reconnectInterval = ymlConfig.ReconnectInterval
	}

	protoVersion := defaultProtoVersion
	if ymlConfig.ProtoVersion > 0 {
		protoVersion = ymlConfig.ProtoVersion
	}

	clientConfig := &gocql.ClusterConfig{
		Hosts:             ymlConfig.Hosts,
		Port:              ymlConfig.Port,
		Timeout:           timeout,
		ConnectTimeout:    connectTimeout,
		ReconnectInterval: reconnectInterval,
		ProtoVersion:      protoVersion,
	}

	cfg := &ClientConfig{ClusterConfig: clientConfig}

	return cfg, nil
}

// CreateSessionFromClientConfigAndKeyspace Creates session from given configuration and keyspace
func CreateSessionFromClientConfigAndKeyspace(config Config, keyspace string) (*gocql.Session, error) {

	gocqlClusterConfig := gocql.NewCluster(HostsAsString(config.Hosts))
	gocqlClusterConfig.Port = config.Port
	gocqlClusterConfig.ConnectTimeout = config.ConnectTimeout
	gocqlClusterConfig.Timeout = config.Timeout
	gocqlClusterConfig.ReconnectInterval = config.ReconnectInterval
	gocqlClusterConfig.ProtoVersion = config.ProtoVersion
	gocqlClusterConfig.Keyspace = keyspace

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
