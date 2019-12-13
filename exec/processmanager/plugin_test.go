// Copyright (c) 2018 Cisco and/or its affiliates.
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

package processmanager_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/cn-infra/v2/exec/processmanager"
	"go.ligato.io/cn-infra/v2/exec/processmanager/status"
	tmpModel "go.ligato.io/cn-infra/v2/exec/processmanager/template/model/process"
)

func TestNewProcess(t *testing.T) {
	RegisterTestingT(t)

	plugin := processmanager.Plugin{}
	plugin.PluginName = "test-pm"
	plugin.PluginDeps.Setup()
	defer func() {
		err := plugin.Close()
		Expect(err).To(BeNil())
	}()

	pr := plugin.NewProcess("name", "command")

	Expect(pr).ToNot(BeNil())
	Expect(pr.GetName()).To(Equal("name"))
	Expect(pr.GetCommand()).To(Equal("command"))
	Expect(pr.GetNotificationChan()).To(BeNil())

	Expect(plugin.GetProcessByName("name")).ToNot(BeNil())
	Expect(plugin.GetAllProcesses()).To(HaveLen(1))

	err := plugin.Delete("name")
	Expect(err).To(BeNil())

	Expect(plugin.GetProcessByName("name")).To(BeNil())
	Expect(plugin.GetAllProcesses()).To(HaveLen(0))
}

func TestNewProcessWithOptions(t *testing.T) {
	RegisterTestingT(t)

	plugin := processmanager.Plugin{}
	plugin.PluginName = "test-pm"
	plugin.PluginDeps.Setup()
	defer func() {
		err := plugin.Close()
		Expect(err).To(BeNil())
	}()

	pr := plugin.NewProcess("name", "command", processmanager.Args("arg1", "arg2"),
		processmanager.Notify(make(chan status.ProcessStatus)))

	Expect(pr).ToNot(BeNil())
	Expect(pr.GetArguments()).To(Equal([]string{"arg1", "arg2"}))
	Expect(pr.GetNotificationChan()).ToNot(BeNil())
}

func TestNewProcessFromTemplate(t *testing.T) {
	RegisterTestingT(t)

	plugin := processmanager.Plugin{}
	plugin.PluginName = "test-pm"
	plugin.PluginDeps.Setup()
	defer func() {
		err := plugin.Close()
		Expect(err).To(BeNil())
	}()

	tmp := &tmpModel.Template{
		Name: "name",
		Cmd:  "command",
		POptions: &tmpModel.TemplatePOptions{
			Args:   []string{"arg1", "arg2"},
			Notify: true,
		},
	}

	pr := plugin.NewProcessFromTemplate(tmp)

	Expect(pr).ToNot(BeNil())
	Expect(pr.GetName()).To(Equal("name"))
	Expect(pr.GetCommand()).To(Equal("command"))
	Expect(pr.GetNotificationChan()).ToNot(BeNil())

	Expect(plugin.GetProcessByName("name")).ToNot(BeNil())
	Expect(plugin.GetAllProcesses()).To(HaveLen(1))

	err := plugin.Delete("name")
	Expect(err).To(BeNil())

	Expect(plugin.GetProcessByName("name")).To(BeNil())
	Expect(plugin.GetAllProcesses()).To(HaveLen(0))
}
