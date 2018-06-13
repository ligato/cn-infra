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

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/namsral/flag"
)

// Agent implements startup & shutdown procedures.
type Agent interface {
	Run() error
	Start() error
	Wait() error
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
	opts Options
}

// Options returns the Options the agent was created with
func (a *agent) Options() Options {
	return a.opts
}

// Start starts the agent.  Start will return as soon as the Agent is ready.  The Agent continues
// running after Start returns.
func (a *agent) Start() error {
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
	return nil
}

// Stop the Agent.  Calls close on all Plugins
func (a *agent) Stop() error {
	// Close plugins
	for _, p := range a.opts.Plugins {
		if err := p.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Wait will not return until a SIGINT, SIGTERM, or SIGKILL is received
// Wait Closes all Plugins before returning
func (a *agent) Wait() error {
	// Wait for signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-sig:
		logrus.DefaultLogger().Info("Signal received, stopping.")
	}

	return a.Stop()
}

// Run runs the agent.  Run will not return until a SIGINT, SIGTERM, or SIGKILL is received
func (a *agent) Run() error {
	if err := a.Start(); err != nil {
		return err
	}
	return a.Wait()
}
