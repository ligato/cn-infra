package main

import (
	"errors"
	"log"

	"github.com/ligato/cn-infra/rpc/rest"
	"golang.org/x/net/context"
	"google.golang.org/grpc/examples/helloworld/helloworld"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/rpc/grpc"
)

// *************************************************************************
// This file contains GRPC service exposure example. To register service use
// Server.RegisterService(descriptor, service)
// ************************************************************************/

const PluginName = "example"

func main() {
	// Init close channel to stop the example after everything was logged
	//exampleFinished := make(chan struct{})

	// Start Agent with ExamplePlugin & FlavorRPC (reused cn-infra plugins).
	/*agent := rpc.NewAgent(rpc.WithPlugins(func(flavor *rpc.FlavorRPC) []*core.NamedPlugin {
		examplePlug := &ExamplePlugin{
			exampleFinished: exampleFinished,
			Deps: Deps{
				PluginLogDeps: *flavor.LogDeps("example"),
				GRPC:          &flavor.GRPC,
			},
		}
		return []*core.NamedPlugin{{examplePlug.PluginName, examplePlug}}
	}))
	core.EventLoopWithInterrupt(agent, exampleFinished)*/

	rest.DefaultPlugin = rest.NewPlugin(
		rest.UseConf(rest.Config{
			Endpoint: ":1234",
		}),
	)

	/*myGRPC := grpc.NewPlugin(
		grpc.UseDeps(grpc.Deps{
			HTTP: myRest,
		}),
	)*/

	p := &ExamplePlugin{
		Deps: Deps{
			Log:  logging.ForPlugin(PluginName),
			GRPC: grpc.DefaultPlugin,
		},
	}

	a := agent.NewAgent(agent.AllPlugins(p))

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

// ExamplePlugin presents main plugin.
type ExamplePlugin struct {
	Deps
}

// Deps are dependencies for ExamplePlugin.
type Deps struct {
	Log  logging.PluginLogger
	GRPC grpc.Server
}

// Init demonstrates the usage of PluginLogger API.
func (plugin *ExamplePlugin) Init() error {
	plugin.Log.Info("Example Init")

	helloworld.RegisterGreeterServer(plugin.GRPC.GetServer(), &GreeterService{})

	return nil
}

func (plugin *ExamplePlugin) Close() error {
	return nil
}

// GreeterService implements GRPC GreeterServer interface (interface generated from protobuf definition file).
// It is a simple implementation for testing/demo only purposes.
type GreeterService struct{}

// SayHello returns error if request.name was not filled otherwise: "hello " + request.Name
func (*GreeterService) SayHello(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	if request.Name == "" {
		return nil, errors.New("not filled name in the request")
	}
	logrus.DefaultLogger().Infof("greeting client: %v", request.Name)

	return &helloworld.HelloReply{Message: "hello " + request.Name}, nil
}
