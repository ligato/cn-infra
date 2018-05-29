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

package main

import (
	"github.com/ligato/cn-infra/wiring"
	"github.com/ligato/cn-infra/examples/wiring/grpc-server/greetingservice"
	"github.com/ligato/cn-infra/rpc/grpc"
)

const (
	defaultHelloWorldGreeting = "hello "
)


func main() {
	closeCh := make(chan struct{})
	// Create a completely empty Plugin
	plugin := &greetingservice.Plugin{}

	// Optionally pass in a config programmatically instead of taking it from grpc.conf or other config file
	//config := &grpc.SetConfig{ Endpoint: "localhost:9111"}
	//plugin.SetConfig(config)

	// wiring.EventLoopWithInterupt will apply the plugins default wiring, create an Agent for it, and run that Agent
	// In the event loop.  Optionally if you wanted to customize the wiring you could have used
	// wiring.EventLoopWithInterrupt(plugin,closeCh,wiring)
	// Or if you wanted to simply customize somewhat the default wiring
	// wiring.EventLoopWithInterrupt(plugin,closeCh,plugin.DefaultWiring(),wiring)
	// which would first apply the DefaultWiring for the plugin, and then custom wiring provided in wiring
	// 99% of the time, you are good with the DefaultWiring
	wiring.EventLoopWithInterrupt(plugin,closeCh,grpc.WithRest(false))
}
