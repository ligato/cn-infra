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

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/onsi/gomega"
)

func Test01DefaultWiring(t *testing.T) {
	gomega.RegisterTestingT(t)
	plugin := &Plugin{}
	err := plugin.Wire(plugin.DefaultWiring(true))
	gomega.Expect(err).Should(gomega.BeNil())
	gomega.Expect(plugin.PluginName).Should(gomega.BeEquivalentTo(defaultName))
	gomega.Expect(plugin.Log).ShouldNot(gomega.BeNil())
	gomega.Expect(plugin.PluginConfig).ShouldNot(gomega.BeNil())
}

func Test02WithNameDefault(t *testing.T) {
	gomega.RegisterTestingT(t)
	plugin := &Plugin{}
	err := plugin.Wire(WithName(true))
	gomega.Expect(err).Should(gomega.BeNil())
	gomega.Expect(plugin.PluginName).Should(gomega.BeEquivalentTo(defaultName))
	gomega.Expect(plugin.Log).Should(gomega.BeNil())
	gomega.Expect(plugin.PluginConfig).Should(gomega.BeNil())
}

func Test03WithNameNonDefault(t *testing.T) {
	gomega.RegisterTestingT(t)
	plugin := &Plugin{}
	name := "foo"
	err := plugin.Wire(WithName(true, name))
	gomega.Expect(err).Should(gomega.BeNil())
	gomega.Expect(plugin.PluginName).Should(gomega.BeEquivalentTo(name))
	gomega.Expect(plugin.Log).Should(gomega.BeNil())
	gomega.Expect(plugin.PluginConfig).Should(gomega.BeNil())
}

func Test04WithLogNonDefault(t *testing.T) {
	gomega.RegisterTestingT(t)
	plugin := &Plugin{}
	log := logging.ForPlugin(defaultName, logrus.NewLogRegistry())
	err := plugin.Wire(WithLog(true, log))
	gomega.Expect(err).Should(gomega.BeNil())
	gomega.Expect(plugin.Log).Should(gomega.BeEquivalentTo(log))
	gomega.Expect(plugin.PluginConfig).Should(gomega.BeNil())
}

func Test05NilWiring(t *testing.T) {
	gomega.RegisterTestingT(t)
	plugin := &Plugin{}
	err := plugin.Wire(nil)
	gomega.Expect(err).Should(gomega.BeNil())
	gomega.Expect(plugin.PluginName).Should(gomega.BeEquivalentTo(defaultName))
	gomega.Expect(plugin.Log).ShouldNot(gomega.BeNil())
	gomega.Expect(plugin.PluginConfig).ShouldNot(gomega.BeNil())
}

func Test06DefaultWiringOverwriteTrue(t *testing.T) {
	gomega.RegisterTestingT(t)
	plugin := &Plugin{}
	name := "foo"

	err := plugin.Wire(WithName(true, name))
	gomega.Expect(err).Should(gomega.BeNil())

	err = plugin.Wire(DefaultWiring(false))
	gomega.Expect(err).Should(gomega.BeNil())
	gomega.Expect(plugin.PluginName).Should(gomega.BeEquivalentTo(defaultName))
	gomega.Expect(plugin.Log).ShouldNot(gomega.BeNil())
	gomega.Expect(plugin.PluginConfig).ShouldNot(gomega.BeNil())

}

func Test07WithNamePrefix(t *testing.T) {
	gomega.RegisterTestingT(t)
	plugin := &Plugin{}
	name := "foo"
	err := plugin.Wire(WithNamePrefix(true, name))
	gomega.Expect(err).Should(gomega.BeNil())
	gomega.Expect(plugin.PluginName).Should(gomega.BeEquivalentTo(name + defaultName))
	gomega.Expect(plugin.Log).Should(gomega.BeNil())
	gomega.Expect(plugin.PluginConfig).Should(gomega.BeNil())
}
