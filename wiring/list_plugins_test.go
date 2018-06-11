package wiring

import (
	"testing"

	"github.com/ligato/cn-infra/core"
	"github.com/onsi/gomega"
)

// TestListPluginsPluginNoDeps checks that the plugin containing no fields
// that would implement Plugin interface (here we miss Close())
// returns empty slice.
func TestListPluginsPluginNoDeps(t *testing.T) {
	gomega.RegisterTestingT(t)

	plugin := &PluginNoDeps{}
	plugs := ListUniqueNamedPlugins(plugin)
	t.Log("plugs ", plugs)
	gomega.Expect(plugs).To(gomega.BeNil())
}

// TestListPluginsOnePluginInDeps checks that the plugin containing multiple fields
// but only one of them implementing Plugin interface (other fields miss Close())
// returns slice with one particular plugin.
func TestListPluginsOnePluginInDeps(t *testing.T) {
	gomega.RegisterTestingT(t)

	plugin := &PluginOneDep{}
	plugs := ListUniqueNamedPlugins(plugin)
	t.Log("plugs ", plugs)
	gomega.Expect(plugs).To(gomega.Equal([]*core.NamedPlugin{{
		core.PluginName("Plugin2"), &plugin.Plugin2}}))
}

// TestListPluginsOnePluginInDeps checks that the plugin containing multiple fields
// but only one of them implementing Plugin interface (other fields miss Close())
// returns slice with one particular plugin.
func TestListPluginsTwoLevelDeps(t *testing.T) {
	gomega.RegisterTestingT(t)

	plugin := &PluginTwoLevelDeps{}
	plugs := ListUniqueNamedPlugins(plugin)
	t.Log("plugs ", plugs)
	gomega.Expect(len(plugs)).To(gomega.Equal(3))
	gomega.Expect(plugs[0]).To(gomega.Equal(&core.NamedPlugin{
		PluginName: core.PluginName("PluginTwoLevelDep1"),
		Plugin:     &plugin.PluginTwoLevelDep1}))
	gomega.Expect(plugs[1]).To(gomega.Equal(&core.NamedPlugin{
		PluginName: core.PluginName("Plugin2"),
		Plugin:     &plugin.PluginTwoLevelDep1.Plugin2}))
	gomega.Expect(plugs[2]).To(gomega.Equal(&core.NamedPlugin{
		PluginName: core.PluginName("PluginTwoLevelDep2"),
		Plugin:     &plugin.PluginTwoLevelDep2}))
}

// PluginNoDeps contains no plugins.
type PluginNoDeps struct {
	Plugin1 MissignCloseMethod
	Plugin2 struct {
		Dep1B string
	}
}

func (*PluginNoDeps) Init() error  { return nil }
func (*PluginNoDeps) Close() error { return nil }

// PluginOneDep contains one plugin (another is missing Close method).
type PluginOneDep struct {
	Plugin1 MissignCloseMethod
	Plugin2 DummyPlugin
}

func (*PluginOneDep) Init() error  { return nil }
func (*PluginOneDep) Close() error { return nil }

type PluginTwoLevelDeps struct {
	PluginTwoLevelDep1 PluginOneDep
	PluginTwoLevelDep2 DummyPlugin
}

func (*PluginTwoLevelDeps) Init() error  { return nil }
func (*PluginTwoLevelDeps) Close() error { return nil }

// MissignCloseMethod implements only Init() but not Close() method.
type MissignCloseMethod struct {
}

// Init does nothing.
func (*MissignCloseMethod) Init() error {
	return nil
}

// DummyPlugin only defines Init() and Close() with empty method bodies.
type DummyPlugin struct {
	internalFlag bool
}

// Init does nothing.
func (*DummyPlugin) Init() error {
	return nil
}

// Close does nothing.
func (*DummyPlugin) Close() error {
	return nil
}

// DummyPlugin only defines Init() and Close() with empty method bodies.
type DummyPluginDep2 struct {
	internalFlag bool
}

// Init does nothing.
func (*DummyPluginDep2) Init() error {
	return nil
}
