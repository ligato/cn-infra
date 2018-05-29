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

package core

import (
	"os"
	"os/signal"
	"syscall"
)

// EventLoopWithInterrupt starts an instance of the agent created with NewAgent().
// Agent is stopped when <closeChan> is closed, a user interrupt (SIGINT), or a
// terminate signal (SIGTERM) is received.
func EventLoopWithInterrupt(agent *Agent, closeCh chan struct{}) error {
	_, errCh := Run(agent, closeCh)
	return <-errCh
}

// Run starts an instance of the agent created with NewAgent().
// Agent is stopped when <closeChan> is closed, , a user interrupt (SIGINT), or a
// terminate signal (SIGTERM) is received.
//
// Returns readyCh which will be closed when the agent is ready - warning readyCh may never be closed if the agent fails on start
// Returns errCh which will receive all errors associated with starting or stopping the event loop
//    errCh will be closed when the agent is successfully stopped
func Run(agent *Agent, closeCh chan struct{}) (<-chan struct{}, <-chan error) {
	// We need two slots in the errCh, one for errors on Start(), one for errors on Stop()
	errCh := make(chan error, 2)
	readyCh := make(chan struct{})
	go func() {
		defer close(errCh)
		if err := agent.Start(); err != nil {
			agent.Error("Error loading core: ", err)
			errCh <- err
			return
		}
		close(readyCh)
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		signal.Notify(sigChan, syscall.SIGTERM)
		select {
		case <-sigChan:
			agent.Println("Interrupt received, returning.")
		case <-closeCh:
		}
		if err := agent.Stop(); err != nil {
			agent.Errorf("Agent stop error '%+v'", err)
			errCh <- err
		}
	}()
	return readyCh, errCh
}
