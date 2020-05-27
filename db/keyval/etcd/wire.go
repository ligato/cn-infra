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

package etcd

import (
	"github.com/google/wire"

	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/db/keyval"
	"go.ligato.io/cn-infra/v2/logging"
)

var WireDefault = wire.NewSet(
	Provider,
	ConfigProvider,
	wire.Struct(new(Deps), "StatusCheck", "Resync", "Serializer"),
	wire.Bind(new(keyval.KvProtoPlugin), new(*Plugin)),
)

func ConfigProvider(conf config.Config) *Config {
	var cfg = DefaultConfig()
	if err := conf.UnmarshalKey("etcd", &cfg); err != nil {
		logging.Errorf("unmarshal key failed: %v", err)
	}
	return cfg
}

func Provider(deps Deps, cfg *Config) (*Plugin, func(), error) {
	p := &Plugin{Deps: deps}
	p.config = cfg
	p.Log = logging.ForPlugin("etcd-client")
	return p, func() {}, nil
}
