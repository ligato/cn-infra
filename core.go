//  Copyright (c) 2020 Cisco and/or its affiliates.
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

package cninfra

import (
	"context"

	"github.com/google/wire"

	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/db/keyval"
	"go.ligato.io/cn-infra/v2/health/probe"
	"go.ligato.io/cn-infra/v2/health/statuscheck"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/logging/logmanager"
	"go.ligato.io/cn-infra/v2/rpc/grpc"
	"go.ligato.io/cn-infra/v2/rpc/prometheus"
	"go.ligato.io/cn-infra/v2/rpc/rest"
	"go.ligato.io/cn-infra/v2/servicelabel"
)

type Base struct {
	ServiceLabel *servicelabel.Plugin // implements: servicelabel.ReaderAPI
	StatusCheck  *statuscheck.Plugin  // implements: statuscheck.PluginStatusWriter, statuscheck.StatusReader

}

func (base Base) Start(ctx context.Context) error {
	logging.Debugf("Base Start()")

	if err := base.ServiceLabel.InitLabel(); err != nil {
		return err
	}
	if err := base.StatusCheck.StartProbing(); err != nil {
		return err
	}

	return nil
}

type Server struct {
	HTTP *rest.Plugin // implements: rest.HTTPHandlers
	GRPC *grpc.Plugin // implements: grpc.Server
}

func (server Server) Start(ctx context.Context) error {
	logging.Debugf("Server Start()")

	if err := server.HTTP.StartServing(); err != nil {
		logging.Fatalf("HTTP serving failed: %v", err)
	}
	if err := server.GRPC.StartServing(); err != nil {
		logging.Fatalf("GRPC serving failed: %v", err)
	}

	return nil
}

type KVStore struct {
	keyval.KvProtoPlugin
}

func ProvideServiceLabelReaderAPI(core Base) servicelabel.ReaderAPI {
	if core.ServiceLabel == nil {
		return nil
	}
	return core.ServiceLabel
}

func ProvideStatusCheckStatusReader(core Base) statuscheck.StatusReader {
	if core.StatusCheck == nil {
		return nil
	}
	return core.StatusCheck
}

func ProvideStatusCheckPluginStatusWriter(core Base) statuscheck.PluginStatusWriter {
	if core.StatusCheck == nil {
		return nil
	}
	return core.StatusCheck
}

var CoreProviders = wire.NewSet(
	ProvideServiceLabelReaderAPI,
	ProvideStatusCheckStatusReader,
	ProvideStatusCheckPluginStatusWriter,
)

func ProvideRestHTTPHandlers(server Server) rest.HTTPHandlers {
	if server.HTTP == nil {
		return nil
	}
	return server.HTTP
}

func ProvideGrpcServer(server Server) grpc.Server {
	if server.GRPC == nil {
		return nil
	}
	return server.GRPC
}

var ServerProviders = wire.NewSet(
	ProvideRestHTTPHandlers,
	ProvideGrpcServer,
)

var WireDefaultCore = wire.NewSet(
	servicelabel.WireDefault,
	statuscheck.WireDefault,
)

var WireDefaultServer = wire.NewSet(
	rest.WireDefault,
	grpc.WireDefault,
)

var WirePrometheusProbe = wire.NewSet(
	probe.WireDefault,
	prometheus.WireDefault,
)

var WireDefaultLogRegistry = wire.NewSet(
	wire.InterfaceValue(new(logging.Registry), logging.DefaultRegistry),
)

var WireDefaultConfig = wire.NewSet(
	wire.InterfaceValue(new(config.Config), config.DefaultConfig),
)

var WireLogManager = wire.NewSet(
	WireDefaultLogRegistry,
	logmanager.WireDefault,
)

/*func StatusCheckProvider(deps statuscheck.Deps, conf *statuscheck.Config) (*statuscheck.Plugin, func(), error) {
	p := &statuscheck.Plugin{Deps: deps}
	p.conf = conf
	p.Log = logging.ForPlugin("status-check")
	cancel := func() {
		if err := p.Close(); err != nil {
			p.Log.Error(err)
		}
	}
	return p, cancel, p.Init()
}

var WireStatusCheck = wire.NewSet(
	StatusCheckProvider,
)*/
