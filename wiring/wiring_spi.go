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
package wiring

import "github.com/ligato/cn-infra/core"

// A Wiring is a function that, when called on a core.Plugin is expected to 'wire'
// Its dependencies
type Wiring func (core.Plugin) error

// A struct is said to be 'Wireable' if it implements a 'Wire' method as specified
// here
type Wireable interface {
	Wire(wiring Wiring) error  // nil Wiring should result in a Default Wiring
}

// A core.Plugin that is also Wireable is a WireablePlugin
type WireablePlugin interface {
	core.Plugin
	Wireable
}

// A thing is said to be Named if it has a method Name() that returns a string
type Named interface  {
	Name() string
}

// A WireablePlugin is Named if it also implements the Named interface
type NamedWirablePlugin interface {
	WireablePlugin
	Named
}

// A struct is said to be DefaultWirable if it has a DefaultWiring() method that returns a Wiring to use to
// Wire up its dependencies with Default values
type DefaultWirable interface {
	DefaultWiring() Wiring
}

// A NamedWireablePlugin that is also DefaultWirable.
type DefaultWireableNamedWireablePlugin interface {
	NamedWirablePlugin
	DefaultWirable
}
