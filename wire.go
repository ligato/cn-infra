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

//+build wireinject

package cninfra

import (
	"context"

	"github.com/google/wire"

	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/datasync/resync"
	"go.ligato.io/cn-infra/v2/health/probe"
	"go.ligato.io/cn-infra/v2/health/statuscheck"
	"go.ligato.io/cn-infra/v2/logging/logmanager"
	"go.ligato.io/cn-infra/v2/rpc/grpc"
	"go.ligato.io/cn-infra/v2/rpc/prometheus"
	"go.ligato.io/cn-infra/v2/rpc/rest"
	"go.ligato.io/cn-infra/v2/servicelabel"
)

//go:generate wire

func InitializeCore(ctx context.Context, conf config.Config) (Core, func(), error) {
	wire.Build(
		WireDefaultLogRegistry,
		logmanager.WireDefault,
		servicelabel.WireDefault,
		statuscheck.WireDefault,
		probe.WireDefault,
		prometheus.WireDefault,
		resync.WireDefault,
		wire.Struct(new(Core), "*"),

		rest.WireDefault,
		grpc.WireDefault,
		wire.Struct(new(Server), "*"),
	)
	return Core{}, nil, nil
}
