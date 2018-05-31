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

package greetingservice

import (
	"errors"
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"github.com/ligato/cn-infra/rpc/grpc"
	"golang.org/x/net/context"
)

const(
	// DefaultHelloWorldGreeting is the string prepended to names passed to the SayHello Service
	DefaultHelloWorldGreeting = "hello "
)

// GreeterService implements GRPC GreeterServer interface (interface generated from protobuf definition file).
// It is a simple implementation for testing/demo only purposes.
type GreeterService struct{}

// SayHello returns error if request.name was not filled otherwise: "hello " + request.Name
func (*GreeterService) SayHello(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	if request.Name == "" {
		return nil, errors.New("not filled name in the request")
	}

	return &helloworld.HelloReply{Message: DefaultHelloWorldGreeting + request.Name}, nil
}

// Plugin creates a new type of Plugin for a particular GRPC Service by embedding grpc.Plugin
type Plugin struct {
	grpc.Plugin
}

// Init overloads the grpc.Plugin Init() method
// It calls grpc.Init() on the embedded gRPC plugin and then register our GreeterService with the GRPC Server
func (plugin *Plugin) Init() (err error) {
	plugin.Log.Infof("Plugin Init()")
	err = plugin.Plugin.Init()
	if err != nil {
		return err
	}
	helloworld.RegisterGreeterServer(plugin.GetServer(), &GreeterService{})
	plugin.Log.Infof("Registered Greeter Service")
	return err
}

// Optionally give our plugin a new name
// func (plugin *Plugin) Name() string { return "greeting-service"}