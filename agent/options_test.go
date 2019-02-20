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

package agent_test

import (
	"testing"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/infra"
	. "github.com/onsi/gomega"
)

func TestDescendantPluginsNoDep(t *testing.T) {
	RegisterTestingT(t)
	plugin := &PluginNoDeps{}
	a := agent.NewAgent(agent.AllPlugins(plugin))
	Expect(a).ToNot(BeNil())
	Expect(a.Options()).ToNot(BeNil())
	Expect(a.Options().Plugins).ToNot(BeNil())
	Expect(len(a.Options().Plugins)).To(Equal(1))
	Expect(a.Options().Plugins[0]).To(Equal(plugin))
}

func TestDescendantPluginsOneLevelDep(t *testing.T) {
	RegisterTestingT(t)

	plugin := &PluginOneDep{}
	plugin.SetName("OneDep")
	a := agent.NewAgent(agent.AllPlugins(plugin))
	Expect(a).ToNot(BeNil())
	Expect(a.Options()).ToNot(BeNil())
	Expect(a.Options().Plugins).ToNot(BeNil())
	Expect(len(a.Options().Plugins)).To(Equal(2))
	Expect(a.Options().Plugins[0]).To(Equal(&plugin.Plugin2))
	Expect(a.Options().Plugins[1]).To(Equal(plugin))
}

func TestDescendantPluginsTwoLevelsDeep(t *testing.T) {
	RegisterTestingT(t)
	plugin := &PluginTwoLevelDeps{}
	plugin.SetName("TwoDep")
	plugin.PluginTwoLevelDep1.SetName("Dep1")
	plugin.PluginTwoLevelDep2.SetName("Dep2")
	a := agent.NewAgent(agent.AllPlugins(plugin))
	Expect(a).ToNot(BeNil())
	Expect(a.Options()).ToNot(BeNil())
	Expect(a.Options().Plugins).ToNot(BeNil())
	Expect(len(a.Options().Plugins)).To(Equal(4))
	Expect(a.Options().Plugins[0]).To(Equal(&plugin.PluginTwoLevelDep1.Plugin2))
	Expect(a.Options().Plugins[1]).To(Equal(&plugin.PluginTwoLevelDep1))
	Expect(a.Options().Plugins[2]).To(Equal(&plugin.PluginTwoLevelDep2))
	Expect(a.Options().Plugins[3]).To(Equal(plugin))

}

func TestDescendantPluginsList(t *testing.T) {
	RegisterTestingT(t)
	plugin := &PluginListDeps{}
	plugin.SetName("ListDep")
	entry1, entry2, entry3 := TestPlugin{}, PluginNoDeps{}, PluginOneDep{}
	entry1.SetName("Dep1")
	entry2.SetName("Dep2")
	entry3.SetName("Dep3")
	entry3.Plugin2.SetName("Dep31")
	plugin.PluginList = append(plugin.PluginList, &entry1, &entry2, &entry3)

	a := agent.NewAgent(agent.AllPlugins(plugin))
	Expect(a).ToNot(BeNil())
	Expect(a.Options()).ToNot(BeNil())
	Expect(a.Options().Plugins).ToNot(BeNil())
	Expect(len(a.Options().Plugins)).To(Equal(5))
	Expect(a.Options().Plugins[0]).To(Equal(&entry1))
	Expect(a.Options().Plugins[1]).To(Equal(&entry2))
	Expect(a.Options().Plugins[2]).To(Equal(&entry3.Plugin2))
	Expect(a.Options().Plugins[3]).To(Equal(&entry3))
	Expect(a.Options().Plugins[4]).To(Equal(plugin))

}

// Various Test Structs after this point

// PluginNoDeps contains no plugins.
type PluginNoDeps struct {
	infra.PluginName
	Plugin1 MissignCloseMethod
	Plugin2 struct {
		Dep1B string
	}
}

func (p *PluginNoDeps) Init() error  { return nil }
func (p *PluginNoDeps) Close() error { return nil }

// PluginOneDep contains one plugin (another is missing Close method).
type PluginOneDep struct {
	infra.PluginName
	Plugin1 MissignCloseMethod
	Plugin2 TestPlugin
}

func (p *PluginOneDep) Init() error  { return nil }
func (p *PluginOneDep) Close() error { return nil }

type PluginTwoLevelDeps struct {
	infra.PluginName
	PluginTwoLevelDep1 PluginOneDep
	PluginTwoLevelDep2 TestPlugin
}

func (p *PluginTwoLevelDeps) Init() error  { return nil }
func (p *PluginTwoLevelDeps) Close() error { return nil }

type TestPlugins = []infra.Plugin

type PluginListDeps struct {
	infra.PluginName
	PluginList TestPlugins
}

func (p *PluginListDeps) Init() error  { return nil }
func (p *PluginListDeps) Close() error { return nil }

// MissignCloseMethod implements only Init() but not Close() method.
type MissignCloseMethod struct {
}

// Init does nothing.
func (*MissignCloseMethod) Init() error {
	return nil
}
