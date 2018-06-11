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

package core

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ligato/cn-infra/logging/logrus"
	. "github.com/onsi/gomega"
)

func TestEmptyAgent(t *testing.T) {
	RegisterTestingT(t)

	agent := NewAgent(Inject(), WithTimeout(1*time.Second))
	Expect(agent).NotTo(BeNil())
	err := agent.Start()
	Expect(err).To(BeNil())
	err = agent.Stop()
	Expect(err).To(BeNil())
}

func TestEventLoopWithInterrupt(t *testing.T) {
	RegisterTestingT(t)

	plugins := []*TestPlugin{{}, {}, {}}

	namedPlugins := []*NamedPlugin{{"First", plugins[0]},
		{"Second", plugins[1]},
		{"Third", plugins[2]}}

	for _, p := range plugins {
		Expect(p.Initialized()).To(BeFalse())
		Expect(p.AfterInitialized()).To(BeFalse())
		Expect(p.Closed()).To(BeFalse())
	}

	agent := NewAgentDeprecated(logrus.DefaultLogger(), 100*time.Millisecond, namedPlugins...)
	closeCh := make(chan struct{})
	errCh := make(chan error)
	go func() {
		errCh <- EventLoopWithInterrupt(agent, closeCh)
	}()

	time.Sleep(100 * time.Millisecond)
	for _, p := range plugins {
		Expect(p.Initialized()).To(BeTrue())
		Expect(p.AfterInitialized()).To(BeTrue())
		Expect(p.Closed()).To(BeFalse())
	}
	close(closeCh)

	select {
	case errCh := <-errCh:
		Expect(errCh).To(BeNil())
	case <-time.After(100 * time.Millisecond):
		t.FailNow()
	}

	for _, p := range plugins {
		Expect(p.Closed()).To(BeTrue())
	}
}

func TestEventLoopFailInit(t *testing.T) {
	RegisterTestingT(t)

	plugins := []*TestPlugin{{}, {}, NewTestPlugin(true, false, false)}

	namedPlugins := []*NamedPlugin{{"First", plugins[0]},
		{"Second", plugins[1]},
		{"Third", plugins[2]}}

	for _, p := range plugins {
		Expect(p.Initialized()).To(BeFalse())
		Expect(p.AfterInitialized()).To(BeFalse())
		Expect(p.Closed()).To(BeFalse())
	}

	agent := NewAgentDeprecated(logrus.DefaultLogger(), 100*time.Millisecond, namedPlugins...)
	closeCh := make(chan struct{})
	errCh := make(chan error)
	go func() {
		errCh <- EventLoopWithInterrupt(agent, closeCh)
	}()

	select {
	case errCh := <-errCh:
		Expect(errCh).NotTo(BeNil())
	case <-time.After(100 * time.Millisecond):
		t.FailNow()
	}

	for _, p := range plugins {
		Expect(p.Initialized()).To(BeTrue())
		// initialization failed of a plugin failed, afterInit was not called
		Expect(p.AfterInitialized()).To(BeFalse())
		Expect(p.Closed()).To(BeTrue())
	}
	close(closeCh)

}

func TestEventLoopAfterInitFailed(t *testing.T) {
	RegisterTestingT(t)

	plugins := []*TestPlugin{{}, NewTestPlugin(false, true, false), {}}

	namedPlugins := []*NamedPlugin{{"First", plugins[0]},
		{"Second", plugins[1]},
		{"Third", plugins[2]}}

	for _, p := range plugins {
		Expect(p.Initialized()).To(BeFalse())
		Expect(p.AfterInitialized()).To(BeFalse())
		Expect(p.Closed()).To(BeFalse())
	}

	agent := NewAgentDeprecated(logrus.DefaultLogger(), 100*time.Millisecond, namedPlugins...)
	closeCh := make(chan struct{})
	errCh := make(chan error)
	go func() {
		errCh <- EventLoopWithInterrupt(agent, closeCh)
	}()

	select {
	case errCh := <-errCh:
		Expect(errCh).NotTo(BeNil())
	case <-time.After(100 * time.Millisecond):
		t.FailNow()
	}

	for _, p := range plugins {
		Expect(p.Initialized()).To(BeTrue())
		Expect(p.Closed()).To(BeTrue())
	}
	close(closeCh)

	Expect(plugins[0].AfterInitialized()).To(BeTrue())
	Expect(plugins[1].AfterInitialized()).To(BeTrue())
	// afterInit of the second plugin failed thus the third was not afterInitialized
	Expect(plugins[2].AfterInitialized()).To(BeFalse())

}

func TestEventLoopCloseFailed(t *testing.T) {
	RegisterTestingT(t)

	plugins := []*TestPlugin{NewTestPlugin(false, false, true), {}, {}}

	namedPlugins := []*NamedPlugin{{"First", plugins[0]},
		{"Second", plugins[1]},
		{"Third", plugins[2]}}

	for _, p := range plugins {
		Expect(p.Initialized()).To(BeFalse())
		Expect(p.AfterInitialized()).To(BeFalse())
		Expect(p.Closed()).To(BeFalse())
	}

	agent := NewAgentDeprecated(logrus.DefaultLogger(), 100*time.Millisecond, namedPlugins...)
	closeCh := make(chan struct{})
	errCh := make(chan error)
	go func() {
		errCh <- EventLoopWithInterrupt(agent, closeCh)
	}()

	time.Sleep(100 * time.Millisecond)
	for _, p := range plugins {
		Expect(p.Initialized()).To(BeTrue())
		Expect(p.AfterInitialized()).To(BeTrue())
		Expect(p.Closed()).To(BeFalse())
	}

	close(closeCh)

	select {
	case errCh := <-errCh:
		Expect(errCh).NotTo(BeNil())
	case <-time.After(100 * time.Millisecond):
		t.FailNow()
	}

	for _, p := range plugins {
		Expect(p.Closed()).To(BeTrue())
	}

}

func TestPluginApi(t *testing.T) {
	RegisterTestingT(t)
	const plName = "Name"
	named := NamedPlugin{plName, &TestPlugin{}}

	strRep := named.String()
	Expect(strRep).To(BeEquivalentTo(plName))
}

type TestPlugin struct {
	failInit      bool
	failAfterInit bool
	failClose     bool

	sync.Mutex
	initCalled      bool
	afterInitCalled bool
	closeCalled     bool
}

func NewTestPlugin(failInit, failAfterInit, failClose bool) *TestPlugin {
	return &TestPlugin{failInit: failInit, failAfterInit: failAfterInit, failClose: failClose}
}

func (p *TestPlugin) Init() error {
	p.Lock()
	defer p.Unlock()
	p.initCalled = true
	if p.failInit {
		return fmt.Errorf("Init failed")
	}
	return nil
}
func (p *TestPlugin) AfterInit() error {
	p.Lock()
	defer p.Unlock()
	p.afterInitCalled = true
	if p.failAfterInit {
		return fmt.Errorf("AfterInit failed")
	}
	return nil
}
func (p *TestPlugin) Close() error {
	p.Lock()
	defer p.Unlock()
	p.closeCalled = true
	if p.failClose {
		return fmt.Errorf("Close failed")
	}
	return nil
}

func (p *TestPlugin) Initialized() bool {
	p.Lock()
	defer p.Unlock()
	return p.initCalled
}

func (p *TestPlugin) AfterInitialized() bool {
	p.Lock()
	defer p.Unlock()
	return p.afterInitCalled
}

func (p *TestPlugin) Closed() bool {
	p.Lock()
	defer p.Unlock()
	return p.closeCalled
}
