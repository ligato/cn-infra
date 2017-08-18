package itest

import (
	"testing"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/generic"
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
}

// Setup registers gomega and starts the agent with the flavor argument
func (t *suiteGenericFlavor) Setup(flavor core.Flavor, golangT *testing.T) {
	t.AgentT.Setup(flavor, t.t)
}

// MockGenericFlavor initializes generic flavor with HTTP mock
func MockGenericFlavor() (*generic.FlavorGeneric, *mock.HTTPMock) {
	httpMock := &mock.HTTPMock{}
	return &generic.FlavorGeneric{
		HTTP: *httpmux.FromExistingServer(httpMock.SetHandler),
	}, httpMock
}

// TC01 asserts that injection works fine and agent starts & stops
func (t *suiteGenericFlavor) TC01StartStop() {
	flavor, _ := MockGenericFlavor()
	t.Setup(flavor, t.T)
	defer t.Teardown()

	gomega.Expect(t.agent).ShouldNot(gomega.BeNil(), "agent is not initialized")
}

// TC03 check that status check in flavor works
func (t *suiteGenericFlavor) TC03StatusCheck() {
	flavor, httpMock := MockGenericFlavor()
	t.Setup(flavor, t.T)
	defer t.Teardown()

	tstPlugin := core.PluginName("tstPlugin")
	flavor.StatusCheck.Register(tstPlugin, nil)
	flavor.StatusCheck.ReportStateChange(tstPlugin, "tst", nil)

	result, err := httpMock.NewRequest("GET", flavor.ServiceLabel.GetAgentPrefix()+
		"/check/status/v1/agent", nil)
	gomega.Expect(err).Should(gomega.BeNil(), "logger is not initialized")
	gomega.Expect(result).ShouldNot(gomega.BeNil(), "http result is not initialized")
	gomega.Expect(result).Should(gomega.BeEquivalentTo(200))
}
