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
	"testing"
	"github.com/ligato/cn-infra/core"
)

func Test07WithRest(t *testing.T) {
	plugin := &Plugin{}
	err := plugin.Wire(nil)
	if err != nil {
		t.Errorf("plugin.Wire returned error %s",err)
	}
	err = plugin.Wire(WithRest(false))
	if err != nil {
		t.Errorf("plugin.Wire returned error %s",err)
	}
	if plugin.PluginName != core.PluginName(defaultName) {
		t.Errorf("Incorrect PluginName, expected %s, actual: %s", defaultName,string(plugin.PluginName))
	}
	if plugin.Log == nil {
		t.Errorf("Incorrect Log, expected non-nil, got nil")
	}
	if plugin.PluginConfig == nil {
		t.Errorf("Incorrect PluginConfig, expected non-nil got nil")
	}
	if plugin.HTTP == nil {
		t.Errorf("Incorrect HTTP, expected non-nil got nil")
	}

}
