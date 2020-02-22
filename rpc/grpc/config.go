// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/namsral/flag"
	"google.golang.org/grpc"

	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/infra"
)

// Config is a configuration for GRPC netListener
// It is meant to be extended with security (TLS...)
type Config struct {
	// Endpoint is an address of GRPC netListener
	Endpoint string `json:"endpoint"`

	// Three or four-digit permission setup for unix domain socket file (if used)
	Permission int `json:"permission"`

	// If set and unix type network is used, the existing socket file will be always removed and re-created
	ForceSocketRemoval bool `json:"force-socket-removal"`

	// Network defaults to "tcp" if unset, and can be set to one of the following values:
	// "tcp", "tcp4", "tcp6", "unix", "unixpacket" or any other value accepted by net.Listen
	Network string `json:"network"`

	// MaxMsgSize returns a ServerOption to set the max message size in bytes for inbound mesages.
	// If this is not set, gRPC uses the default 4MB.
	MaxMsgSize int `json:"max-msg-size"`

	// MaxConcurrentStreams returns a ServerOption that will apply a limit on the number
	// of concurrent streams to each ServerTransport.
	MaxConcurrentStreams uint32 `json:"max-concurrent-streams"`

	// TLS info:
	InsecureTransport bool     `json:"insecure-transport"`
	Certfile          string   `json:"cert-file"`
	Keyfile           string   `json:"key-file"`
	CAfiles           []string `json:"ca-files"`

	// ExtendedLogging enables detailed GRPC logging
	ExtendedLogging bool `json:"extended-logging"`

	// PrometheusMetrics enables prometheus metrics for gRPC client.
	PrometheusMetrics bool `json:"prometheus-metrics"`

	// Compression for inbound/outbound messages.
	// Supported only gzip.
	//TODO Compression string
}

func (cfg *Config) getGrpcOptions() (opts []grpc.ServerOption) {
	if cfg.MaxConcurrentStreams > 0 {
		opts = append(opts, grpc.MaxConcurrentStreams(cfg.MaxConcurrentStreams))
	}
	if cfg.MaxMsgSize > 0 {
		opts = append(opts, grpc.MaxRecvMsgSize(cfg.MaxMsgSize))
	}
	return
}

func (cfg *Config) getTLS() (*tls.Config, error) {
	// Check if explicitly disabled.
	if cfg.InsecureTransport {
		return nil, nil
	}
	// Minimal requirement is to get cert and key for enabling TLS.
	if cfg.Certfile == "" && cfg.Keyfile == "" {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(cfg.Certfile, cfg.Keyfile)
	if err != nil {
		return nil, err
	}
	tc := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
	}

	// Check if we want verify client's certificate against custom CA
	if len(cfg.CAfiles) > 0 {
		caCertPool := x509.NewCertPool()
		for _, c := range cfg.CAfiles {
			cert, err := ioutil.ReadFile(c)
			if err != nil {
				return nil, err
			}

			ok := caCertPool.AppendCertsFromPEM(cert)
			if !ok {
				return nil, fmt.Errorf("unable to add CA from '%s' file", c)
			}
		}
		tc.ClientCAs = caCertPool
		tc.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tc, nil
}

func (cfg *Config) getSocketType() string {
	// Default to tcp socket type of not specified for backward compatibility
	if cfg.Network == "" {
		return "tcp"
	}
	return cfg.Network
}

// GetPort parses suffix from endpoint & returns integer after last ":" (otherwise it returns 0)
func (cfg *Config) GetPort() int {
	if cfg.Endpoint != "" && cfg.Endpoint != ":" {
		index := strings.LastIndex(cfg.Endpoint, ":")
		if index >= 0 {
			port, err := strconv.Atoi(cfg.Endpoint[index+1:])
			if err == nil {
				return port
			}
		}
	}

	return 0
}

// DeclareGRPCPortFlag declares GRPC port (with usage & default value) a flag for a particular plugin name
func DeclareGRPCPortFlag(pluginName infra.PluginName) {
	plugNameUpper := strings.ToUpper(string(pluginName))

	usage := "Configure Agent' " + plugNameUpper + " net listener (port & timeouts); also set via '" +
		plugNameUpper + config.EnvSuffix + "' env variable."
	flag.String(grpcPortFlag(pluginName), "", usage)
}

func grpcPortFlag(pluginName infra.PluginName) string {
	return strings.ToLower(string(pluginName)) + "-port"
}
