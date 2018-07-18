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
	"io"
	"net/http"
	"strconv"

	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/unrolled/render"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

// Plugin maintains the GRPC netListener (see Init, AfterInit, Close methods)
type Plugin struct {
	Deps

	*Config

	option Option
	//Option func(*Plugin)

	// GRPC server instance
	grpcServer *grpc.Server
	// Used mainly for testing purposes
	listenAndServe ListenAndServe
	// GRPC network listener
	netListener io.Closer
	// Plugin availability flag
	disabled bool
}

// Deps is a list of injected dependencies of the GRPC plugin.
type Deps struct {
	core.PluginName
	Log          logging.PluginLogger
	PluginConfig config.PluginConfig
	HTTP         rest.HTTPHandlers
}

// Init prepares GRPC netListener for registration of individual service
func (plugin *Plugin) Init() (err error) {
	// Get GRPC configuration file
	if plugin.Config == nil {
		plugin.Config, err = plugin.getGrpcConfig()
		if err != nil || plugin.disabled {
			return err
		}
	}

	// Prepare GRPC server
	if plugin.grpcServer == nil {
		opts := plugin.Config.getGrpcOptions()
		plugin.grpcServer = grpc.NewServer(opts...)
		grpclog.SetLogger(plugin.Log.NewLogger("grpc-server"))
	}

	// Start GRPC listener
	if plugin.listenAndServe != nil {
		plugin.netListener, err = plugin.listenAndServe(*plugin.Config, plugin.grpcServer)
	} else {
		plugin.Log.Info("Listening GRPC on tcp://", plugin.Config.Endpoint)
		plugin.netListener, err = ListenAndServeGRPC(plugin.Config, plugin.grpcServer)
	}

	return err
}

// AfterInit starts the HTTP netListener.
func (plugin *Plugin) AfterInit() (err error) {
	//plugin.Log.Debugf("GRPC AfterInit()")

	if plugin.Deps.HTTP != nil {
		plugin.Log.Infof("exposing GRPC services via HTTP (port %v) on: /service",
			strconv.Itoa(plugin.Deps.HTTP.GetPort()))
		plugin.Deps.HTTP.RegisterHTTPHandler("/service", func(formatter *render.Render) http.HandlerFunc {
			return plugin.grpcServer.ServeHTTP
		}, "GET", "PUT", "POST")
	} else {
		plugin.Log.Infof("HTTP not set, skip exposing GRPC services")
	}

	return err
}

// Close stops the HTTP netListener.
func (plugin *Plugin) Close() error {
	wasError := safeclose.Close(plugin.netListener)

	if plugin.grpcServer != nil {
		plugin.grpcServer.Stop()
	}

	return wasError
}

// GetServer is a getter for accessing grpc.Server
func (plugin *Plugin) GetServer() *grpc.Server {
	return plugin.grpcServer
}

// IsDisabled returns *true* if the plugin is not in use due to missing
// grpc configuration.
func (plugin *Plugin) IsDisabled() (disabled bool) {
	return plugin.disabled
}

func (plugin *Plugin) getGrpcConfig() (*Config, error) {
	var grpcCfg Config
	found, err := plugin.PluginConfig.GetValue(&grpcCfg)
	if err != nil {
		return &grpcCfg, err
	}
	if !found {
		plugin.Log.Info("GRPC config not found, skip loading this plugin")
		plugin.disabled = true
	}
	return &grpcCfg, nil
}
