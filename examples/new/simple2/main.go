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
	"fmt"
	"log"
	"net/http"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/examples/new/greeter"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/rpc/grpc"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/unrolled/render"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

func main() {
	// Prepare plugin
	//restPlugin := rest.NewPlugin()
	//restPlugin := rest.DefaultPlugin
	restPlugin := rest.NewPlugin(rest.UseConf(rest.Config{Endpoint: "asd"}))

	myPlugin := &MyPlugin{
		REST: restPlugin,
		GRPC: grpc.NewPlugin(

			grpc.UseConf(grpc.Config{
				Endpoint: "0.0.0.0:9111",
			}),
		),
	}

	// Create agent
	agent := agent.NewAgent(agent.DescendantPlugins(myPlugin))

	// Run agent
	if err := agent.Run(); err != nil {
		log.Fatal(err)
	}
}

// MyPlugin wraps grpc.Plugin and rest.Plugin register the GreeterServer as part of its Init()
type MyPlugin struct {
	REST *rest.Plugin
	GRPC *grpc.Plugin
}

// Init registers GreeterServer with the embedded grpc.Plugin and sets up REST
func (plugin *MyPlugin) Init() error {
	logrus.DefaultLogger().Infof("MyPlugin Init()")

	helloworld.RegisterGreeterServer(plugin.GRPC.GetServer(), new(greeter.Service))

	plugin.REST.RegisterHTTPHandler("/mine", func(formatter *render.Render) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hey")
		}
	}, "GET")

	logrus.DefaultLogger().Infof("Registered Greeter Service")

	return nil
}

// Close plugin
func (plugin *MyPlugin) Close() error {
	return nil
}
