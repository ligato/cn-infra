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

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/unrolled/render"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/cn-infra/v2/logging/logrus"
	"go.ligato.io/cn-infra/v2/rpc/rest"
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

	tlsConfig  *tls.Config
	auther     *Authenticator
	serverOpts []grpc.ServerOption
	metrics    *grpc_prometheus.ServerMetrics
	limiter    *rate.Limiter
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
		if err != nil {
			return err
		}
	}
	if p.Config.Disabled {
		p.Log.Infof("grpc server disabled via config")
		return nil
	}

	// Prepare GRPC server
	if p.grpcServer == nil {
		// If config for TLS was not provided with the `UseTLS` option, check config file.
		if p.tlsConfig == nil {
			tc, err := p.Config.getTLS()
			if err != nil {
				return err
			}
			p.tlsConfig = tc
		}

		// use default grpc prometheus metrics if not set by UseServerMetrics option.
		if p.metrics == nil && p.Config.PrometheusMetrics {
			p.metrics = grpc_prometheus.DefaultServerMetrics
		}

		var unaryChain []grpc.UnaryServerInterceptor
		var streamChain []grpc.StreamServerInterceptor

		// Rate limiting middleware
		if p.limiter != nil {
			p.Log.Debugf("Rate limiter set to rate %.1f req/s (%d max burst)", p.limiter.Limit(), p.limiter.Burst())
			unaryChain = append(unaryChain, UnaryServerInterceptorLimiter(p.limiter))
			streamChain = append(streamChain, StreamServerInterceptorLimiter(p.limiter))
		}

		// Auth middleware
		if p.auther != nil {
			p.Log.Debug("Token authentication for gRPC enabled")
			unaryChain = append(unaryChain, grpc_auth.UnaryServerInterceptor(p.auther.Authenticate))
			streamChain = append(streamChain, grpc_auth.StreamServerInterceptor(p.auther.Authenticate))

		}

		// add server options for prometheus metrics
		if p.metrics != nil {
			p.Log.Debug("Prometheus server metrics for gRPC enabled")
			unaryChain = append(unaryChain, p.metrics.UnaryServerInterceptor())
			streamChain = append(streamChain, p.metrics.StreamServerInterceptor())
		}

		// get server options from config
		opts := p.Config.getGrpcOptions()

		opts = append(opts,
			grpc_middleware.WithUnaryServerChain(unaryChain...),
			grpc_middleware.WithStreamServerChain(streamChain...),
		)

		// add custom server options
		opts = append(opts, p.serverOpts...)

		if p.tlsConfig != nil {
			p.Log.Debug("Secure connection (TLS) for gRPC enabled")
			opts = append(opts, grpc.Creds(credentials.NewTLS(p.tlsConfig)))
		}

		p.grpcServer = grpc.NewServer(opts...)
	}

	grpcLogger := logrus.NewLogger("grpc-server")
	if p.Config != nil && p.Config.ExtendedLogging {
		p.Log.Debug("GRPC transport logging enabled")
		grpcLogger.SetVerbosity(logLevel)
	}
	grpclog.SetLoggerV2(grpcLogger)

	if p.Deps.HTTP != nil {
		p.Log.Infof("exposing GRPC services via HTTP (port %v) on: /service", p.Deps.HTTP.GetPort())
		p.Deps.HTTP.RegisterHTTPHandler("/service", func(formatter *render.Render) http.HandlerFunc {
			return p.grpcServer.ServeHTTP
		}, "GET", "PUT", "POST")
	} else {
		p.Log.Debugf("HTTP not set, skip exposing GRPC services")
	}

	return nil
}

// AfterInit starts the HTTP netListener.
func (p *Plugin) AfterInit() (err error) {
	if p.Config.Disabled {
		return nil
	}

	err = p.StartServing()
	if err != nil {
		return err
	}

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
	return p.Config.Disabled
}

func (p *Plugin) getGrpcConfig() (*Config, error) {
	grpcCfg := DefaultConfig()
	found, err := p.Cfg.LoadValue(grpcCfg)
	if err != nil {
		return grpcCfg, err
	}
	if !found {
		p.Log.Infof("GRPC config not found, using default config: %+v", grpcCfg)
	}
	return grpcCfg, nil
}

func (p *Plugin) StartServing() (err error) {
	// initialize prometheus metrics for grpc server
	if p.metrics != nil {
		p.metrics.InitializeMetrics(p.grpcServer)
	}

	// Start GRPC listener
	p.netListener, err = ListenAndServe(p.Config, p.grpcServer)
	if err != nil {
		return err
	}
	p.Log.Infof("Listening GRPC on: %v", p.Config.Endpoint)

	return nil
}
