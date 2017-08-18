package itest

import (
	"testing"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/generic"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/httpmux"
	"github.com/ligato/cn-infra/httpmux/mock"
	"github.com/onsi/gomega"
)

type suiteGenericFlavor struct {
	T *testing.T
	AgentT
	Given
	When
	Then
	mock.HttpMock
}

func MockGenericFlavor(mock *mock.HttpMock) *generic.FlavorGeneric {
	return &generic.FlavorGeneric{
		HTTP: *httpmux.FromExistingServer(mock.SetHandler),
	}
}

// TC01 asserts that injection works fine and agent starts & stops
func (t *suiteGenericFlavor) TC01StartStop() {
	flavor := MockGenericFlavor(&t.HttpMock)
	t.Setup(flavor, t.T)

	gomega.Expect(t.agent).ShouldNot(gomega.BeNil(), "agent is not initialized")

	defer t.Teardown()
}

// TC03 check that status check in flavor works
func (t *suiteGenericFlavor) TC03StatusCheck() {
	flavor := &local.FlavorLocal{}
	t.Setup(flavor, t.T)

	tstPlugin := core.PluginName("tstPlugin")
	flavor.StatusCheck.Register(tstPlugin, nil)
	flavor.StatusCheck.ReportStateChange(tstPlugin, "tst", nil)

	t.HttpMock.NewRequest("GET", flavor.ServiceLabel.GetAgentPrefix()+
		"/check/status/v1/agent", nil)
	//TODO assert flavor.StatusCheck using IDX map???

	defer t.Teardown()
}
