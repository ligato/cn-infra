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

package agent

import (
	"os"
	"os/signal"
	"syscall"

	"errors"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/utils/once"
	"github.com/namsral/flag"
)

// Agent implements startup & shutdown procedures.
type Agent interface {
	Run() error
	Start() error
	Wait() error
	After() <-chan struct{}
	Error() error
	Stop() error
	Options() Options
}

// NewAgent creates a new Agent
func NewAgent(opts ...Option) Agent {
	options := newOptions(opts...)

	flag.Parse()

	return &agent{
		opts: options,
	}
}

type agent struct {
	opts      Options
	stopCh    chan struct{}
	startOnce once.ReturnError
	stopOnce  once.ReturnError
}

// Options returns the Options the agent was created with
func (a *agent) Options() Options {
	return a.opts
}

// Start starts the agent.  Start will return as soon as the Agent is ready.  The Agent continues
// running after Start returns.
func (a *agent) Start() error {
	return a.startOnce.Do(a.startSignalWrapper)
}

func (a *agent) startSignalWrapper() error {
	// If we want to properly handle cleanup when a SIG comes in *during*
	// agent startup (ie, clean up after its finished) we need to register
	// for the signal before we start() the agent
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	err := a.start()
	// If the agent started, we have things to clean up if here is a SIG
	// So fire off a goroutine to do that
	if err == nil {
		go func() {
			// Wait for signal or agent stop
			select {
			case <-sig:
			case <-a.After():
			}
			logrus.DefaultLogger().Info("Signal received, stopping.")
			// Doesn't hurt to call Stop twice, its idempotent because of the
			// stopOnce
			a.Stop()
			signal.Stop(sig)
			close(sig)
		}()
	}
	return err
}

func (a *agent) start() error {
	// Init plugins
	for _, p := range a.opts.Plugins {
		if err := p.Init(); err != nil {
			return err
		}
	}
	// AfterInit plugins
	for _, p := range a.opts.Plugins {
		var plug core.Plugin = p
		if np, ok := p.(*core.NamedPlugin); ok {
			plug = np.Plugin
		}
		if postPlugin, ok := plug.(core.PostInit); ok {
			if err := postPlugin.AfterInit(); err != nil {
				return err
			}
		} else {
			logrus.DefaultLogger().Debugf("plugin %v has no AfterInit", p)
		}
	}
	a.stopCh = make(chan struct{}, 1) // If we are started, we have a stopCh to signal stopping
	logrus.DefaultLogger().Info("Agent Started")
	return nil
}

// Stop the Agent.  Calls close on all Plugins
func (a *agent) Stop() error {
	return a.stopOnce.Do(a.stop)
}

func (a *agent) stop() error {
	if a.stopCh != nil { // Don't stop if we didn't start
		defer close(a.stopCh)
		// Close plugins
		for _, p := range a.opts.Plugins {
			if err := p.Close(); err != nil {
				return err
			}
		}
		logrus.DefaultLogger().Info("Agent Stopped.")
		return nil
	}
	err := errors.New("attempted to stop an agent that wasn't Started")
	logrus.DefaultLogger().Error(err)
	return err

}

// Wait will not return until a SIGINT, SIGTERM, or SIGKILL is received
// Or the Agent is Stopped
// All Plugins are Closed() before Wait returns
func (a *agent) Wait() error {
	if a.stopCh != nil { // Don't wait if we didn't Start
		<-a.After()
		// If we get here, a.Stop() has already been called, and we are simply
		// retrieving the error if any squirreled away by stopOnce
		return a.Stop()
	}
	err := errors.New("attempted to wait on an agent that wasn't Started")
	logrus.DefaultLogger().Error(err)
	return err
}

// Run runs the agent.  Run will not return until a SIGINT, SIGTERM, or SIGKILL is received
func (a *agent) Run() error {
	if err := a.Start(); err != nil {
		return err
	}
	return a.Wait()
}

// After returns a channel that will be closed when the agent is Stopped.
// To retrieve any error from the agent stopping call Error() on the agent
// The normal pattern of use is:
//
// agent := NewAgent(options...)
// agent.Start()
// select {
// case <-agent.After() // Will wait till the agent is stopped
// ...
// }
// err := agent.Error() // Will return any error from the agent being stopped
//
func (a *agent) After() <-chan struct{} {
	if a.stopCh != nil {
		return a.stopCh
	}
	// The agent didn't start, so we can't return a.stopCh
	// because *only* a.start() should allocate that
	// we won't return a nil channel, because nil channels
	// block forever.
	// Since the normal pattern is to call a.After() so you
	// can select till the agent is done and a.Stop() to
	// retrieve the error, returning a closed channel will preserve that
	// usage, as a.Stop() returns an error complaining that the agent
	// never started.
	stopCh := make(chan struct{})
	close(stopCh)
	return stopCh
}

// Error returns any error that occurred when the agent was Stopped
func (a *agent) Error() error {
	// a.Stop() returns whatever error occurred when stopping the agent
	// This is because of stopOnce
	// If you try to retrieve an error before the agent is started, you will get
	// an error complaining the agent isn't started.
	return a.Stop()
}
