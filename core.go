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
	"github.com/google/wire"

	"go.ligato.io/cn-infra/v2/datasync/resync"
	"go.ligato.io/cn-infra/v2/health/probe"
	"go.ligato.io/cn-infra/v2/health/statuscheck"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/logging/logmanager"
	"go.ligato.io/cn-infra/v2/rpc/grpc"
	"go.ligato.io/cn-infra/v2/rpc/prometheus"
	"go.ligato.io/cn-infra/v2/rpc/rest"
	"go.ligato.io/cn-infra/v2/servicelabel"
)

var WireDefaultLogRegistry = wire.NewSet(
	wire.InterfaceValue(new(logging.Registry), logging.DefaultRegistry),
)

type Core struct {
	LogRegistry  logging.Registry
	LogManager   *logmanager.Plugin
	ServiceLabel *servicelabel.Plugin //servicelabel.ReaderAPI

	StatusCheck *statuscheck.Plugin //statuscheck.PluginStatusWriter
	Probe       *probe.Plugin
	Prometheus  *prometheus.Plugin

	Resync *resync.Plugin

	Server
}
type Server struct {
	HTTP *rest.Plugin //rest.HTTPHandlers
	GRPC *grpc.Plugin //grpc.Server
}

func (core Core) Run() {
	logging.Debugf("Core Run()")

	core.StatusCheck.StartProbing()

	if err := core.HTTP.StartServing(); err != nil {
		panic(err)
	}

	if err := core.GRPC.StartServing(); err != nil {
		panic(err)
	}
}

func RestHTTPHandlersProvider(server Core) rest.HTTPHandlers {
	if server.HTTP == nil {
		return nil
	}
	return server.HTTP
}

func GrpcServerProvider(server Core) grpc.Server {
	if server.GRPC == nil {
		return nil
	}
	return server.GRPC
}
