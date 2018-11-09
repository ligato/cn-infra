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

package process

// POptions is common object which holds all selected options
type POptions struct {
	args []string
	count int
	detach bool
}

// POption is helper function to set process options
type POption func(*POptions)

// Args if process should start with arguments
func Args(args ...string) POption {
	return func(p *POptions) {
		p.args = args
	}
}

// Restarts defines number of automatic restarts of given process
func Restarts(count int) POption {
	return func(p *POptions) {
		p.count = count
	}
}

// Detach process from parent after start, so it can survive after parent process is terminated
func Detach() POption {
	return func(p *POptions) {
		p.detach = true
	}
}