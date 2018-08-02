package main

import (
	"log"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/bolt"
	"github.com/ligato/cn-infra/logging"
)

func main() {
	p := &ExamplePlugin{
		Log:             logging.ForPlugin("example"),
		DB:              &bolt.DefaultPlugin,
		exampleFinished: make(chan struct{}),
	}

	a := agent.NewAgent(
		agent.AllPlugins(p),
		agent.QuitOnClose(p.exampleFinished),
	)
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

// ExamplePlugin demonstrates the usage of Bolt plugin.
type ExamplePlugin struct {
	Log logging.PluginLogger
	DB  keyval.KvProtoPlugin

	exampleFinished chan struct{}
}

// Init demonstrates using Bolt plugin.
func (p *ExamplePlugin) Init() (err error) {
	db := p.DB.NewBroker(keyval.Root)

	// Store some data
	txn := db.NewTxn()
	txn.Put("/agent/config/interface/iface0", nil)
	txn.Put("/agent/config/interface/iface1", nil)
	txn.Commit()

	// List keys
	const listPrefix = "/agent/config/interface/"

	p.Log.Infof("List BoltDB keys: %s", listPrefix)

	keys, err := db.ListKeys(listPrefix)
	if err != nil {
		p.Log.Fatal(err)
	}

	for {
		key, val, all := keys.GetNext()
		if all == true {
			break
		}

		p.Log.Infof("Key: %q Val: %v", key, val)
	}

	return nil
}

// AfterInit closes the example.
func (p *ExamplePlugin) AfterInit() (err error) {
	close(p.exampleFinished)
	return nil
}

// Close frees plugin resources.
func (p *ExamplePlugin) Close() error {
	return nil
}

// String returns name of plugin.
func (p *ExamplePlugin) String() string {
	return "example"
}
