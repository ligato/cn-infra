//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package grpc

import (
	"crypto/tls"
	"fmt"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"

	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/rpc/rest"
)

// DefaultPlugin is a default instance of Plugin.
var DefaultPlugin = *NewPlugin()

// NewPlugin creates a new Plugin with the provided Options.
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	p.PluginName = "grpc"
	//p.HTTP= &rest.DefaultPlugin // turned off by default

	for _, o := range opts {
		o(p)
	}

	if p.Deps.Log == nil {
		p.Deps.Log = logging.ForPlugin(p.String())
	}
	if p.Deps.Cfg == nil {
		p.Deps.Cfg = config.ForPlugin(p.String(),
			config.WithExtraFlags(func(flags *config.FlagSet) {
				flags.String(grpcPortFlag(p.PluginName), "", fmt.Sprintf(
					"Configure %q server port", p.String()))
			}))
	}

	return p
}

// Option is a function that can be used in NewPlugin to customize Plugin.
type Option func(*Plugin)

// UseConf returns Option which injects a particular configuration.
func UseConf(conf Config) Option {
	return func(p *Plugin) {
		p.Config = &conf
	}
}

// UseDeps returns Option that can inject custom dependencies.
func UseDeps(cb func(*Deps)) Option {
	return func(p *Plugin) {
		cb(&p.Deps)
	}
}

// UseHTTP returns Option that sets HTTP handlers.
func UseHTTP(h rest.HTTPHandlers) Option {
	return func(p *Plugin) {
		p.Deps.HTTP = h
	}
}

// UseAuth returns Option that sets Authenticator.
func UseAuth(a *Authenticator) Option {
	return func(p *Plugin) {
		p.auther = a
	}
}

// UseTLS returns Option that sets TLS config.
func UseTLS(c *tls.Config) Option {
	return func(p *Plugin) {
		p.tlsConfig = c
	}
}

// UseServerOpts returns Option that adds server options.
func UseServerOpts(o ...grpc.ServerOption) Option {
	return func(p *Plugin) {
		p.serverOpts = append(p.serverOpts, o...)
	}
}

// UsePromMetrics returns Option that sets custom server metrics.
func UsePromMetrics(metrics *grpc_prometheus.ServerMetrics) Option {
	return func(p *Plugin) {
		p.metrics = metrics
	}
}

// UseRateLimiter returns an Option which sets rate limiter.
func UseRateLimiter(r *rate.Limiter) Option {
	return func(p *Plugin) {
		p.limiter = r
	}
}
