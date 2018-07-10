package main

import (
	"log"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
)

// *************************************************************************
// This file contains logger use cases. To define a custom logger, use
// PluginLogger.NewLogger(name). The logger is using 6 levels of logging:
// - Debug
// - Info (this one is default)
// - Warn
// - Error
// - Panic
// - Fatal
//
// Global log levels can be changed locally with the Logger.SetLevel()
// or remotely using REST (but different flavor must be used: rpc.RpcFlavor).
// ************************************************************************/

const PluginName = "example"

func main() {
	// Init close channel to stop the example after everything was logged
	exampleFinished := make(chan struct{})

	// Start Agent with ExamplePlugin & LocalFlavor (reused cn-infra plugins).
	/*agent := local.NewAgent(local.WithPlugins(func(flavor *local.FlavorLocal) []*core.NamedPlugin {
		examplePlug := &ExamplePlugin{
			exampleFinished: exampleFinished,
			PluginLogDeps:   *flavor.LogDeps("logs-example"),
		}
		return []*core.NamedPlugin{{examplePlug.PluginName, examplePlug}}
	}))
	core.EventLoopWithInterrupt(agent, exampleFinished)*/

	p := &ExamplePlugin{
		Log:             logging.ForPlugin(PluginName, logrus.DefaultRegistry),
		exampleFinished: exampleFinished,
	}
	a := agent.NewAgent(
		agent.AllPlugins(p),
		agent.QuitOn(exampleFinished),
	)
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

// ExamplePlugin presents the PluginLogger API.
type ExamplePlugin struct {
	Log logging.PluginLogger
	//local.PluginLogDeps

	exampleFinished chan struct{}
}

// Init demonstrates the usage of PluginLogger API.
func (plugin *ExamplePlugin) Init() (err error) {
	exampleString := "example"
	exampleNum := 15

	// Set log level which logs only entries with current severity or above
	plugin.Log.SetLevel(logging.WarnLevel)  // warn, error, panic, fatal
	plugin.Log.SetLevel(logging.InfoLevel)  // info, warn, error, panic, fatal - default log level
	plugin.Log.SetLevel(logging.DebugLevel) // everything

	// Basic logger options
	plugin.Log.Print("----------- Log examples -----------")
	plugin.Log.Printf("Print with format specifier. String: %s, Digit: %d, Value: %v", exampleString, exampleNum, plugin)

	// Format also available for all 6 levels of log levels
	plugin.Log.Debug("Debug log example: Debugging information")
	plugin.Log.Info("Info log example: Something informative")
	plugin.Log.Warn("Warn log example: Something unexpected, warning")
	plugin.Log.Error("Error log example: Failure without exit")
	plugin.showPanicLog()
	//log.Fatal("Bye") calls os.Exit(1) after logging

	// Log with field - automatically adds timestamp
	plugin.Log.WithField("exampleString", exampleString).Info("Info log with field example")
	// For multiple fields
	plugin.Log.WithFields(map[string]interface{}{"exampleString": exampleString, "exampleNum": exampleNum}).Info("Info log with field example string and num")

	// Custom (child) logger with name
	childLogger := plugin.Log.NewLogger("childLogger")
	// Usage of custom loggers
	childLogger.Infof("Log using named logger with name: %v", childLogger.GetName())
	childLogger.Debug("Debug log using childLogger!")

	childLogger2 := plugin.Log.NewLogger("childLogger2")
	childLogger2.Debug("Debug log using childLogger2!")

	// End the example
	plugin.Log.Info("logs in plugin example finished, sending shutdown ...")

	close(plugin.exampleFinished)

	return nil
}

// Close implements Plugin interface..
func (plugin *ExamplePlugin) Close() (err error) {
	return nil
}

// Name implements PluginNamed interface.
func (plugin *ExamplePlugin) Name() string {
	return PluginName
}

// showPanicLog demonstrates panic log + recovering.
func (plugin *ExamplePlugin) showPanicLog() {
	defer func() {
		if err := recover(); err != nil {
			plugin.Log.Info("Recovered from panic")
		}
	}()
	plugin.Log.Panic("Panic log: calls panic() after log, will be recovered") //calls panic() after logging
}
