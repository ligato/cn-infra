package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/examples/new/greeter"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/rpc/grpc"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/unrolled/render"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

func main() {
	// Using default deps
	/*plugin := &MyPlugin{Plugin: grpc.NewPlugin()}
	 */

	// Using custom deps
	/*plugin := &MyPlugin{Plugin: grpc.NewPlugin(
		grpc.UsingDeps(grpc.Deps{
			//PluginName: "myGRPC",
		}),
	)}*/

	// Using custom config
	/*plugin := &MyPlugin{
		grpc.NewPlugin(grpc.UsingConf(grpc.Config{
			Endpoint: "0.0.0.0:9111",
		})),
	}*/

	// Using custom deps
	/*plugin := &MyPlugin{
		Log: logging.ForPlugin("myPlugin", logrus.DefaultRegistry),
		Plugin: grpc.NewPlugin(
			grpc.UsingDeps(grpc.Deps{
				HTTP: rest.NewPlugin(),
			}),
			grpc.UsingConf(grpc.Config{
				Endpoint: "0.0.0.0:9111",
			}),
		),
	}*/

	// Using various options
	restPlugin := rest.NewPlugin(
		rest.UseConf(rest.Config{
			Endpoint:       "0.0.0.0:9191",
			ServerCertfile: "server.crt",
			ServerKeyfile:  "server.key",
			ClientCerts:    []string{"ca.crt"},
		}),
	)
	grpcPlugin := grpc.NewPlugin(
		grpc.UseDeps(grpc.Deps{
			HTTP: restPlugin,
		}),
		grpc.UseConf(grpc.Config{
			Endpoint: "0.0.0.0:9111",
		}),
	)
	myPlugin := &MyPlugin{
		Log:    logging.ForPlugin("myPlugin", logrus.DefaultRegistry),
		Plugin: grpcPlugin,
	}

	// Create agent
	agent := agent.NewAgent(
		agent.Version("1.0.0"),
		/*agent.MaxStartupTime(time.Second*5),
		 */
		agent.Plugins(
			/*core.NamePlugin("myPlugin", &MyPlugin{grpc.NewPlugin()}),
			 */
			restPlugin,
			grpcPlugin,
			core.NamePlugin("myPlugin", myPlugin),
		),
	)

	// Initialize flags
	//agent.Init()

	// Run agent
	if err := agent.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// MyPlugin aggregates a Log and a grpc.Plugin
type MyPlugin struct {
	Log logging.PluginLogger
	*grpc.Plugin
}

// Init plugin
func (plugin *MyPlugin) Init() (err error) {
	plugin.Log.Infof("MyPlugin Init()")

	plugin.Deps.HTTP.RegisterHTTPHandler("/mine", func(formatter *render.Render) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hey")
		}
	}, "GET")

	helloworld.RegisterGreeterServer(plugin.GetServer(), new(greeter.Service))

	plugin.Log.Infof("Registered Greeter Service")
	return err
}
