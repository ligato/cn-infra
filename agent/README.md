# Agent [![GoDoc](https://godoc.org/github.com/ligato/cn-infra/agent?status.svg)](https://godoc.org/github.com/ligato/cn-infra/agent)

The **agent** package provides life-cycle managment agent for plugins.
It intented tp be used as a base point of your program used in main package.

```go
func main() {
	plugin := myplugin.NewPlugin()
	
	a := agent.NewAgent(agent.Plugins(plugin))
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
```