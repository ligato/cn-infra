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

package logmanager

import (
	"github.com/google/wire"

	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/logging"
)

var WireDefault = wire.NewSet(
	ConfigProvider,
	Provider,
	wire.Struct(new(Deps), "ServiceLabel", "LogRegistry", "HTTP"),
)

func ConfigProvider(conf config.Config) *Config {
	var cfg = DefaultConfig()
	if err := conf.UnmarshalKey("logs", &cfg); err != nil {
		logging.Errorf("unmarshal key failed: %v", err)
	}
	return cfg
}

func Provider(deps Deps, conf *Config) (*Plugin, error) {
	p := new(Plugin)
	p.Deps = deps
	p.config = conf
	p.Log = logging.ForPlugin("logmanager")
	return p, p.Init()
}
