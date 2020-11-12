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

package config_test

import (
	"os"
	"testing"
	"time"

	flag "github.com/spf13/pflag"

	. "github.com/onsi/gomega"

	"go.ligato.io/cn-infra/v2/config"
)

func TestFlagName(t *testing.T) {
	RegisterTestingT(t)

	const pluginName = "testplugin"

	Expect(config.FlagName(pluginName)).To(BeEquivalentTo("testplugin-config"))
}

type testConf struct {
	Label   string
	Details string
	Port    int
	Rate    float64
	Debug   bool
	Timeout time.Duration
	List    []string
	Map     map[string]string
	Section *struct {
		Name string
	}
}

func TestForPluginWithConfigFile(t *testing.T) {
	RegisterTestingT(t)

	const pluginName = "testplugin"
	const configFileName = pluginName + ".conf"

	setConfigDir(t, "./testdata")

	pluginConfig := config.ForPlugin(pluginName)
	Expect(pluginConfig).ShouldNot(BeNil())

	Expect(flag.CommandLine.Lookup(config.FlagName(pluginName))).To(BeNil())
	//config.DefineFlagsFor(pluginName)
	flag.CommandLine.AddFlagSet(config.GetFlagSetFor(pluginName))

	Expect(pluginConfig.GetConfigName()).Should(BeEquivalentTo(configFileName))
	Expect(flag.CommandLine.Lookup(config.FlagName(pluginName))).ToNot(BeNil())

	cfg := testConf{
		Label:   "no_label",
		Details: "NONE",
	}
	Expect(pluginConfig.LoadValue(&cfg)).To(BeTrue())
	Expect(cfg.Label).To(Equal("bar"))
	Expect(cfg.Details).To(Equal("NONE"))
	Expect(cfg.Port).To(Equal(5))
	Expect(cfg.Rate).To(Equal(2.1))
	Expect(cfg.Debug).To(Equal(true))
	Expect(cfg.Timeout).To(Equal(time.Second * 3))
	Expect(cfg.List).To(ConsistOf("a", "b", "c"))
	Expect(cfg.Map).To(And(
		HaveKeyWithValue("a", "A"),
		HaveKeyWithValue("b", "B"),
	))
	Expect(cfg.Section.Name).To(Equal("xyz"))
}

func TestForPluginWithoutConfigFile(t *testing.T) {
	RegisterTestingT(t)

	const pluginName = "confignofileplugin"

	setConfigDir(t, "./testdata")

	pluginConfig := config.ForPlugin(pluginName)
	Expect(pluginConfig).ShouldNot(BeNil())

	//config.DefineFlagsFor(pluginName)
	flag.CommandLine.AddFlagSet(config.GetFlagSetFor(pluginName))
	Expect(pluginConfig.GetConfigName()).Should(BeEmpty())
}

func TestForPluginWithSpecifiedConfigFile(t *testing.T) {
	RegisterTestingT(t)

	const pluginName = "myplugin"
	const configFileName = "testplugin.conf"

	setConfigDir(t, "./testdata")

	pluginConfig := config.ForPlugin(pluginName,
		config.WithCustomizedFlag(config.FlagName(pluginName), configFileName, "customized config filename"),
	)
	Expect(pluginConfig).ShouldNot(BeNil())

	//config.DefineFlagsFor(pluginName)
	flag.CommandLine.AddFlagSet(config.GetFlagSetFor(pluginName))
	Expect(pluginConfig.GetConfigName()).Should(BeEquivalentTo(configFileName))
}

func setConfigDir(t *testing.T, dir string) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			panic(err)
		}
	})
}
