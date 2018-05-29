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
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"github.com/ligato/cn-infra/wiring"
	"github.com/ligato/cn-infra/rpc/grpc"
	"github.com/ligato/cn-infra/examples/wiring/grpc-server/greetingservice"
	"golang.org/x/net/context"
	"testing"
	dialer "google.golang.org/grpc"
	"time"
	"net"
)

const (
	defaultAddress = "localhost:9111"
	defaultHelloWorldName = "Ed"
)


func Setup(config *grpc.Config) (plugin *greetingservice.Plugin,closeCh chan struct{},readyCh chan interface{},errorCh chan error ) {
	closeCh = make(chan struct{})
	readyCh = make(chan interface {})
	errorCh = make(chan error)
	plugin = &greetingservice.Plugin{}
	plugin.SetConfig(config)
	go func() { errorCh <- wiring.MonitorableEventLoopWithInterupt(plugin,closeCh,readyCh);close(errorCh)}()
	return plugin,closeCh,readyCh,errorCh
}

func GreetingTest(t *testing.T,config *grpc.Config,name string) {
	plugin,closeCh,readyCh,errorCh := Setup(config)
	defer close(closeCh)

	select {
	case err := <-errorCh:
		t.Errorf("%s failed with err on running EventLoop: %s", name, err)
		return
	case <- readyCh:
	}

	plugin.Log.Infof("Attempting to dial target %s", config.Endpoint)
	d := func(target string, duration time.Duration) (net.Conn, error) {
		network := "tcp"
		if config.SocketType != "" {
			network = config.SocketType
		}
		return net.DialTimeout(network,target,duration)
	}
	conn, err := dialer.Dial(config.Endpoint, dialer.WithInsecure(), dialer.WithDialer(d))
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
	expectedResponse := &helloworld.HelloReply{Message:(greetingservice.DefaultHelloWorldGreeting + defaultHelloWorldName)}
	if r.Message != expectedResponse.Message {
		t.Errorf("%s failed with incorrect response to SayHello on GRPC Server \"%s\" expected \"%s\"",name,r.Message,expectedResponse.Message)
		return
	}
}

func TestGrpc01TCP (t *testing.T) {
	GreetingTest(t,&grpc.Config{ Endpoint: defaultAddress},"TestGrpcTCP01")
}

//TODO: Fix this test (or the code)
//func TestGrpc01UnixFileSocket (t *testing.T) {
//	name := "TestGrpcUnixFileSocket01"
//	tempfile,err := ioutil.TempFile("/tmp","test")
//	defer tempfile.Close()
//	defer os.Remove(tempfile.Name())
//	if err != nil {
//		t.Errorf("%s failed to open tempfile, %s",name,err)
//	}
//	GreetingTest(t,&grpc.SetConfig{ Endpoint: tempfile.Name(),SocketType:"unix"},name)
//}