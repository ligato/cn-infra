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
	"io"
	"net/http"
	"strconv"

	"google.golang.org/grpc/grpclog"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/unrolled/render"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// set according to google.golang.org/grpc/internal/transport/log.go
// transport package only logs to verbose level 2 by default
const logLevel = 2

// Plugin maintains the GRPC netListener (see Init, AfterInit, Close methods)
type Plugin struct {
	Deps

	*Config

	// GRPC server instance
	grpcServer *grpc.Server
	// GRPC network listener
	netListener io.Closer
	// Plugin availability flag
	disabled bool

	tlsConfig *tls.Config
	auther    *Authenticator
}

// Deps is a list of injected dependencies of the GRPC plugin.
type Deps struct {
	infra.PluginDeps
	HTTP rest.HTTPHandlers
}

// Init prepares GRPC netListener for registration of individual service
func (p *Plugin) Init() (err error) {
	// Get GRPC configuration file
	if p.Config == nil {
		p.Config, err = p.getGrpcConfig()
		if err != nil || p.disabled {
			return err
		}
	}

	// Prepare GRPC server
	if p.grpcServer == nil {
		opts := p.Config.getGrpcOptions()

		// If config for TLS was not provided with the `UseTLS` option, check config file.
		if p.tlsConfig == nil {
			tc, err := p.Config.getTLS()
			if err != nil {
				return err
			}
			p.tlsConfig = tc
		}

		if p.tlsConfig != nil {
			p.Log.Info("Secure connection for gRPC enabled")
			opts = append(opts, grpc.Creds(credentials.NewTLS(p.tlsConfig)))
		}

		if p.auther != nil {
			p.Log.Info("Token authentication for gRPC enabled")
			opts = append(opts, grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(p.auther.Authenticate)))
			opts = append(opts, grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(p.auther.Authenticate)))
		}
		p.grpcServer = grpc.NewServer(opts...)
	}

	grpcLogger := p.Log.NewLogger("grpc-server")
	if p.Config != nil && p.Config.ExtendedLogging {
		p.Log.Info("GRPC transport logging enabled")
		grpcLogger.SetVerbosity(logLevel)
	}
	grpclog.SetLoggerV2(grpcLogger)

	return nil
}

// AfterInit starts the HTTP netListener.
func (p *Plugin) AfterInit() (err error) {
	if p.disabled {
		return nil
	}

	if p.Deps.HTTP != nil {
		p.Log.Infof("exposing GRPC services via HTTP (port %v) on: /service",
			strconv.Itoa(p.Deps.HTTP.GetPort()))
		p.Deps.HTTP.RegisterHTTPHandler("/service", func(formatter *render.Render) http.HandlerFunc {
			return p.grpcServer.ServeHTTP
		}, "GET", "PUT", "POST")
	} else {
		p.Log.Infof("HTTP not set, skip exposing GRPC services")
	}

	// Start GRPC listener
	p.netListener, err = ListenAndServe(p.Config, p.grpcServer)
	if err != nil {
		return err
	}
	p.Log.Infof("Listening GRPC on: %v", p.Config.Endpoint)

	return nil
}

// Close stops the HTTP netListener.
func (p *Plugin) Close() error {
	if p.grpcServer != nil {
		p.grpcServer.Stop()
	}
	return nil
}

// GetServer is a getter for accessing grpc.Server
func (p *Plugin) GetServer() *grpc.Server {
	return p.grpcServer
}

// IsDisabled returns *true* if the plugin is not in use due to missing
// grpc configuration.
func (p *Plugin) IsDisabled() bool {
	return p.disabled
}

func (p *Plugin) getGrpcConfig() (*Config, error) {
	var grpcCfg Config
	found, err := p.Cfg.LoadValue(&grpcCfg)
	if err != nil {
		return &grpcCfg, err
	}
	if !found {
		p.Log.Info("GRPC config not found, skip loading this plugin")
		p.disabled = true
	}
	return &grpcCfg, nil
}
