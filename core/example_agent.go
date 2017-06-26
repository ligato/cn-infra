package core

import (
	"os"
	"os/signal"
)

// ExampleAgent struct with public channel used to close it
type ExampleAgent struct {
	CloseChannel chan *struct{}
}

// EventLoopWithInterrupt init Agent with plugins. Agent can be interrupted from outside using public CloseChannel
func (exampleAgent *ExampleAgent) EventLoopWithInterrupt(agent *Agent) {
	err := agent.Start()
	if err != nil {
		agent.log.Error("Error loading core", err)
		os.Exit(1)
	}
	defer func() {
		err := agent.Stop()
		if err != nil {
			agent.log.Errorf("Agent stop error '%+v'", err)
			os.Exit(1)
		}
	}()

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt)
	select {
	case <-sigChan:
		agent.log.Info("Interrupt received, returning.")
		return
	case <-exampleAgent.CloseChannel:
		err := agent.Stop()
		if err != nil {
			agent.log.Errorf("Agent stop error '%v'", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
}
