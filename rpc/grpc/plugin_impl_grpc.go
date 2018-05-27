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
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/wiring"
	"github.com/pkg/errors"
)

const (
	defaultName = "grpc"
)

// Plugin maintains the GRPC netListener (see Init, AfterInit, Close methods)
type Plugin struct {
	Deps
	// Stored GRPC config (used in example)
	grpcCfg *Config
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
	Log        logging.PluginLogger
	PluginName core.PluginName
	HTTP       rest.HTTPHandlers
	config.PluginConfig
}

// Init prepares GRPC netListener for registration of individual service
func (plugin *Plugin) Init() error {
	var err error
	// Get GRPC configuration file
	if plugin.grpcCfg == nil {
		plugin.grpcCfg, err = plugin.getGrpcConfig()
		if err != nil || plugin.disabled {
			return err
		}
	}

	// Prepare GRPC server
	if plugin.grpcServer == nil {
		var opts []grpc.ServerOption
		if plugin.grpcCfg.MaxConcurrentStreams > 0 {
			opts = append(opts, grpc.MaxConcurrentStreams(plugin.grpcCfg.MaxConcurrentStreams))
		}
		if plugin.grpcCfg.MaxMsgSize > 0 {
			opts = append(opts, grpc.MaxMsgSize(plugin.grpcCfg.MaxMsgSize))
		}

		plugin.grpcServer = grpc.NewServer(opts...)
		grpclog.SetLogger(plugin.Log.NewLogger("grpc-server"))
	}

	// Start GRPC listener
	if plugin.listenAndServe != nil {
		plugin.netListener, err = plugin.listenAndServe(*plugin.grpcCfg, plugin.grpcServer)
	} else {
		plugin.Log.Info("Listening GRPC on tcp://", plugin.grpcCfg.Endpoint)
		plugin.netListener, err = ListenAndServeGRPC(plugin.grpcCfg, plugin.grpcServer)
	}

	return err
}

// AfterInit starts the HTTP netListener.
func (plugin *Plugin) AfterInit() (err error) {
	if plugin.Deps.HTTP != nil {
		plugin.Log.Info("exposing GRPC services over HTTP port " + strconv.Itoa(plugin.Deps.HTTP.GetPort()) +
			" /service ")
		plugin.Deps.HTTP.RegisterHTTPHandler("service", func(formatter *render.Render) http.HandlerFunc {
			return plugin.grpcServer.ServeHTTP
		}, "GET", "PUT", "POST")
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

// String returns plugin name (if not set defaults to "HTTP")
func (plugin *Plugin) String() string {
	if plugin.Deps.PluginName != "" {
		return string(plugin.Deps.PluginName)
	}
	return "GRPC"
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

// Implements wiring.Wireable This allows us to use Wiring rather than Flavors
// to configure Plugin Dependencies.
func (plugin *Plugin) Wire(wiring wiring.Wiring) error {
	if wiring == nil {
		wiring = plugin.DefaultWiring()
	}
	return wiring(plugin)
}

// Implements wiring.DefaultWirable .  Allows us to get a fully wired version of this file
// without having to specify any wiring.
func (plugin *Plugin) DefaultWiring() wiring.Wiring {
	flavor := &local.FlavorLocal{}
	flavor.Inject()
	return plugin.DefaultWiringFromFlavorLocal(flavor)
}

// Abuse slightly the old flavor way of doing things.  Rather than having to figure out all of the
// Wiring myself at this stage, steal from the local.FlavorLocal.  This makes transitioning a Plugin
// to being wirable very easy, as you can simply copy paste from a Flavor, as all Flavor's start
// life with local.FlavorLocal.
// Note: this default wiring doesn't configuring Deps.HTTP, because its designed to Wire the Plugin
// standalone.  If you *wanted* to configure Deps.HTTP you could simply supply a Wiring that does.
func (plugin *Plugin) DefaultWiringFromFlavorLocal(flavor *local.FlavorLocal) wiring.Wiring {
	ret := func (plugin core.Plugin) error {
		p,ok := plugin.(*Plugin)
		if ok {
			grpcPlugDeps := *flavor.InfraDeps(defaultName, local.WithConf())
			p.Deps.Log = grpcPlugDeps.Log
			//p.Deps.Log = logging.ForPlugin("grpc",nil)
			p.Deps.PluginConfig = grpcPlugDeps.PluginConfig
			p.Deps.PluginName = grpcPlugDeps.PluginName
			p.Deps.HTTP = nil;
			return nil;
		} else {
			return errors.New("grpc.DefaultWiringFromFlavorLocal could not convert core.Plugin to *grpc.Plugin")
		}
	}
	return ret;
}

// Implement wiring.Named
func (plugin *Plugin) Name() string {
	return string(plugin.Deps.PluginName)
}

// Convienence function to set the runtime configuration of the GRPC plugin rather than reading it from file
func (plugin *Plugin) Config(config *Config) {
	plugin.grpcCfg = config
}

