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

package main

import (
	"log"
	"time"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/examples/new/greeter"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/rpc/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

func main() {
	// Create agent
	a := agent.NewAgent(
		agent.Version("1.0.0"),
		agent.MaxStartupTime(time.Second*5),
		agent.Recursive(
			&MyPlugin{grpc.NewPlugin(
				grpc.UseConf(grpc.Config{
					Endpoint: "localhost:9191",
				}),
				grpc.UseDeps(grpc.Deps{
					PluginName: "myGRPC",
				}),
			)},
		),
	)

	// Run agent
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

type MyPlugin struct {
	GRPC *grpc.Plugin
}

func (plugin *MyPlugin) Init() error {
	logrus.DefaultLogger().Infof("MyPlugin Init()")

	helloworld.RegisterGreeterServer(plugin.GRPC.GetServer(), new(greeter.GreeterService))

	logrus.DefaultLogger().Infof("Registered Greeter Service")

	return nil
}

func (plugin *MyPlugin) Close() error {
	return nil
}
