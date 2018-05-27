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
package grpc_test

import (
	"errors"
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"github.com/ligato/cn-infra/rpc/grpc"
	"github.com/ligato/cn-infra/wiring"
	"golang.org/x/net/context"
	"testing"
	dialer "google.golang.org/grpc"
)

const (
	defaultAddress = "localhost:9111"
	defaultHelloWorldGreeting = "hello "
	defaultHelloWorldName = "Ed"
)

// GreeterService implements GRPC GreeterServer interface (interface generated from protobuf definition file).
// It is a simple implementation for testing/demo only purposes.
type GreeterService struct{}

// SayHello returns error if request.name was not filled otherwise: "hello " + request.Name
func (*GreeterService) SayHello(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	if request.Name == "" {
		return nil, errors.New("not filled name in the request")
	}

	return &helloworld.HelloReply{Message: defaultHelloWorldGreeting + request.Name}, nil
}

type GreeterServicePlugin struct {
	grpc.Plugin
}

func (plugin *GreeterServicePlugin) Init() (err error) {
	plugin.Log.Infof("GreeterServicePlugin Init()")
	err = plugin.Plugin.Init()
	if err != nil {
		return err
	} else {
		helloworld.RegisterGreeterServer(plugin.GetServer(), &GreeterService{})
		plugin.Log.Infof("Registered Greeter Service")
	}
	return err
}

func (plugin *GreeterServicePlugin) Name() string { return "greeting-service"}

func GreetingTest(t *testing.T,config *grpc.Config,name string, target string) {
	closeCh := make(chan struct{})
	defer close(closeCh)
	readyCh := make(chan interface {})
	errorCh := make(chan error)
	plugin := &GreeterServicePlugin{}
	plugin.Config(config)
	go func() { errorCh <- wiring.MonitorableEventLoopWithInterupt(plugin,closeCh,readyCh);close(errorCh)}()
	var err error
	select {
	case err = <-errorCh:
		t.Errorf("%s failed with err on running EventLoop: %s", name, err)
		return
	case <- readyCh:
	}
	plugin.Log.Infof("Attempting to dial target %s", target)
	conn, err := dialer.Dial(target, dialer.WithInsecure())
	if err != nil {
		t.Errorf("%s failed with err on Dialing GRPC Server: %s", name, err)
		return
	}
	defer conn.Close()
	c := helloworld.NewGreeterClient(conn)
	r, err := c.SayHello(context.Background(), &helloworld.HelloRequest{Name: defaultHelloWorldName})
	if err != nil {
		t.Errorf("%s failed with err on calling SayHello on GRPC Server %s",name,err)
		return
	}
	expectedResponse := &helloworld.HelloReply{Message:(defaultHelloWorldGreeting + defaultHelloWorldName)}
	if r.Message != expectedResponse.Message {
		t.Errorf("%s failed with incorrect response to SayHello on GRPC Server \"%s\" expected \"%s\"",name,r.Message,expectedResponse.Message)
		return
	}
}

func TestGrpc01TCP (t *testing.T) {
	GreetingTest(t,&grpc.Config{ Endpoint: defaultAddress},"TestGrpcTCP01",defaultAddress)
}

// TODO: Fix this test (or the code)
//func TestGrpc01UnixFileSocket (t *testing.T) {
//	name := "TestGrpcUnixFileSocket01"
//	tempfile,err := ioutil.TempFile("/tmp","test")
//	defer tempfile.Close()
//	defer os.Remove(tempfile.Name())
//	if err != nil {
//		t.Errorf("%s failed to open tempfile, %s",name,err)
//	}
//	GreetingTest(t,&grpc.Config{ UnixSocketFilePath: tempfile.Name()},name,("unix://"+tempfile.Name()))
//}