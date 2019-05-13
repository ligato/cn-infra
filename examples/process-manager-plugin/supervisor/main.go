package main

import (
	"fmt"
	"time"

	"github.com/ligato/cn-infra/agent"

	bs "github.com/ligato/cn-infra/exec/supervisor"
	"github.com/ligato/cn-infra/logging"
)

func main() {
	// The supervisor plugin defines a configuration file allowing to easily manage processes using process
	// manager plugin.
	//
	// The config file can be put to the supervisor either via flag "-supervisor-config="
	// or define its path in the environment variable "SUPERVISOR_CONFIG". Another option is
	// to define config directly with "UseConf()" - this option will be shown in the example. A sample of
	// YAML config file can be found in the processes/supervisor folder.

	log := logging.DefaultLogger

	// Define three processes p1-p3 where the third process is set to terminate p2 if stopped
	conf := bs.Config{
		Processes: []bs.Process{
			{
				Name:        "p1",
				LogFilePath: "example.log",
				BinaryPath:  "../test-process/test-process",
				Args:        []string{"-max-uptime=60"},
			},
			{
				Name:        "p2",
				LogFilePath: "example.log",
				BinaryPath:  "../test-process/test-process",
			},
			{
				Name:        "p3",
				LogFilePath: "example.log",
				BinaryPath:  "../test-process/test-process",
				Required:    true,
			},
		},
	}

	// start plugin
	bsp := bs.NewPlugin(bs.UseConf(conf))

	a := agent.NewAgent(agent.AllPlugins(bsp))
	if err := a.Start(); err != nil {
		panic(err)
	}
	defer func() {
		if err := a.Stop(); err != nil {
			panic(err)
		}
	}()

	// give the agent time to start
	time.Sleep(5 * time.Second)

	// test if all processes are running
	checkRunning("p1", bsp, log)
	checkRunning("p2", bsp, log)
	checkRunning("p3", bsp, log)

	// terminate p1
	stopProcess("p1", bsp, log)

	// test if all states are as required
	checkStopped("p1", bsp, log)
	checkRunning("p2", bsp, log)
	checkRunning("p3", bsp, log)

	// terminate p3 (should also terminate p2)
	log.Info("The p3 is going to be terminated which should stop the p2 as well")
	stopProcess("p3", bsp, log)
}

func checkRunning(name string, bsp bs.Supervisor, log logging.Logger) {
	p1 := bsp.GetProcessByName(name)
	if p1 == nil {
		panic(fmt.Sprintf("expected running process %s", name))
	}
	log.Infof("process %s is running", name)
}

func checkStopped(name string, bsp bs.Supervisor, log logging.Logger) {
	p1 := bsp.GetProcessByName(name)
	if p1 != nil {
		panic(fmt.Sprintf("expected stopped process %s", name))
	}
	log.Infof("process %s is stopped", name)
}

func stopProcess(name string, bsp bs.Supervisor, log logging.Logger) {
	p1 := bsp.GetProcessByName(name)
	if p1 == nil {
		panic(fmt.Sprintf("expected running process %s", name))
	}
	if _, err := p1.StopAndWait(); err != nil {
		panic(fmt.Sprintf("failed to stop process %s: %v", name, err))
	}

	// give the process time to stop
	time.Sleep(2 * time.Second)

	log.Infof("process %s stopped", name)
}
