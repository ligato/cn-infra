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

package pluginutils_test

import (
	"testing"

	"github.com/ligato/cn-infra/utils/pluginutils"
	. "github.com/onsi/gomega"
)

//func TestFilter(t *testing.T) {
//	gomega.RegisterTestingT(t)
//	plugins := []core.Plugin{&Plugin{}, &Plugin{name: "foo"}, &Plugin{}, &Plugin{}}
//	plugin.Filter(plugins, func(plugin core.Plugin) bool {
//		return plugin.name == "foo"
//	})
//}

// TestDescendantsNoDeps checks that the plugin containing no fields
// that would implement Plugin interface (here we miss Close())
// returns empty slice.
func TestWalkNoDeps(t *testing.T) {
	RegisterTestingT(t)

	p := &Plugin{}
	pl := &pluginutils.List{}
	err := pluginutils.Walk(p, pl.WalkFunc)
	Expect(err).To(BeNil())
	Expect(len(pl.Plugins)).To(Equal(1))
	Expect(pl.Plugins[0]).To(Equal(p))
}

// TestDescendantsInDeps checks that the plugin containing multiple fields
// but only one of them implementing Plugin interface (other fields miss Close())
// returns slice with one particular plugin.
func TestWalkOneLevel(t *testing.T) {
	RegisterTestingT(t)

	p := &PluginOneDep{}
	pl := pluginutils.List{}
	err := pluginutils.Walk(p, pl.WalkFunc)
	Expect(err).To(BeNil())
	Expect(len(pl.Plugins)).To(Equal(2))
	Expect(pl.Plugins[0]).To(Equal(p))
	Expect(pl.Plugins[1]).To(Equal(&p.Plugin2))
}

// TestDescendantsTwoLevelDeps checks that the plugin containing multiple fields
// but only one of them implementing Plugin interface (other fields miss Close())
// returns slice with one particular plugin.
func TestWalkTwoLevel(t *testing.T) {
	RegisterTestingT(t)

	p := &PluginTwoLevel{}
	pl := pluginutils.List{}
	err := pluginutils.Walk(p, pl.WalkFunc)
	Expect(err).To(BeNil())
	Expect(len(pl.Plugins)).To(Equal(4))
	Expect(pl.Plugins[0]).To(Equal(p))
	Expect(pl.Plugins[1]).Should(Equal(&p.PluginTwoLevelDep1))
	Expect(pl.Plugins[2]).Should(Equal(&p.PluginTwoLevelDep2))
	Expect(pl.Plugins[3]).Should(Equal(&p.PluginTwoLevelDep1.Plugin2))
}

// Plugin contains no plugins.
type Plugin struct {
	Plugin1 NotAPlugin
	Plugin2 struct {
		Dep1B string
	}
}

func (*Plugin) Init() error  { return nil }
func (*Plugin) Close() error { return nil }

// PluginOneDep contains one plugin (another is missing Close method).
type PluginOneDep struct {
	DepOneLevel
}

type DepOneLevel struct {
	Plugin1 NotAPlugin
	Plugin2 Plugin
}

func (*PluginOneDep) Init() error  { return nil }
func (*PluginOneDep) Close() error { return nil }

type PluginTwoLevel struct {
	PluginTwoLevelDep1 PluginOneDep
	PluginTwoLevelDep2 Plugin
}

func (*PluginTwoLevel) Init() error  { return nil }
func (*PluginTwoLevel) Close() error { return nil }

// NotAPlugin implements only Init() but not Close() method.
type NotAPlugin struct {
}

// Init does nothing.
func (*NotAPlugin) Init() error {
	return nil
}
