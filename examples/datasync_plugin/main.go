package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/db/keyval/kvproto"
	"github.com/ligato/cn-infra/examples/model"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/config"
	"golang.org/x/net/context"
	"github.com/ligato/cn-infra/flavors/localdeps"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/namsral/flag"
)

// *************************************************************************
// This file contains examples of simple Data Broker CRUD operations
// (APIs) including an event handler (watcher). The CRUD operations
// supported by the publisher (data broker) are as follows:
// - Create/Update: publisher.Put()
// - Read:          publisher.Get()
// - Delete:        publisher.Delete()
//
// These functions are called from the REST API. CRUD operations are
// done as single operations and as a part of the transaction
// ************************************************************************/

/********
 * Main *
 ********/

var log logging.Logger

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized. Agent is instantiated with generic plugins (ETCD, Kafka, Status check,
// HTTP and Log), resync plugin and example plugin which demonstrates ETCD functionality.
func main() {
	log = logroot.StandardLogger()
	// Init close channel to stop the example
	closeChannel := make(chan struct{}, 1)

	flavor := ExampleFlavor{}

	// Create new agent
	agent := core.NewAgent(log, 15*time.Second, append(flavor.Plugins())...)

	// End when the ETCD example is finished
	go closeExample("etcd txn example finished", closeChannel)

	core.EventLoopWithInterrupt(agent, closeChannel)
}

// Stop the agent with desired info message
func closeExample(message string, closeChannel chan struct{}) {
	time.Sleep(12 * time.Second)
	log.Info(message)
	closeChannel <- struct{}{}
}

/**********
 * Flavor *
 **********/

// define ETCD flag to load config
func init() {
	flag.String("etcdv3-config", "etcd.conf",
		"Location of the Etcd configuration file")
}

type ExampleFlavor struct {
	// Local flavor to access to Infra (logger, service label, status check)
	Local local.FlavorLocal
	// Etcd plugin
	ETCD         etcdv3.Plugin
	// Etcd sync which manages and injects connection
	ETCDDataSync kvdbsync.Plugin
	// Example plugin
	DatasyncExample ExamplePlugin

	injected bool
}

// Inject sets object references
func (ef *ExampleFlavor) Inject() (allReadyInjected bool) {
	// Init local flavor
	ef.Local.Inject()
	// Init ETCD + ETCD sync
	ef.ETCD.Deps.PluginInfraDeps = *ef.Local.InfraDeps("etcdv3")
	ef.ETCDDataSync.Deps.PluginLogDeps = *ef.Local.LogDeps("etcdv3-datasync")
	ef.ETCDDataSync.KvPlugin = &ef.ETCD
	ef.ETCDDataSync.ResyncOrch = &ef.Local.ResyncOrch
	ef.ETCDDataSync.ServiceLabel = &ef.Local.ServiceLabel
	// Inject infra + transport (publisher, watcher) to example plugin
	ef.DatasyncExample.InfraDeps = *ef.Local.InfraDeps("datasync-example")
	ef.DatasyncExample.Publisher = &ef.ETCDDataSync
	ef.DatasyncExample.Watcher = &ef.ETCDDataSync

	return true
}

// Plugins combines all Plugins in flavor to the list
func (ef *ExampleFlavor) Plugins() []*core.NamedPlugin {
	ef.Inject()
	return core.ListPluginsInFlavor(ef)
}

/**********************
 * Example plugin API *
 **********************/

// PluginID of the custom ETCD plugin
const PluginID core.PluginName = "example-plugin"

/******************
 * Example plugin *
 ******************/

// ExamplePlugin implements Plugin interface which is used to pass custom plugin instances to the agent
type ExamplePlugin struct {
	Deps

	exampleConfigurator *ExampleConfigurator        // Plugin configurator
	Publisher           datasync.KeyProtoValWriter  // To write ETCD data
	Watcher             datasync.KeyValProtoWatcher // To watch ETCD data
	changeChannel       chan datasync.ChangeEvent   // Channel used by the watcher for change events
	resyncChannel       chan datasync.ResyncEvent   // Channel used by the watcher for resync events
	context             context.Context             // Used to cancel watching
	watchDataReg        datasync.WatchRegistration  // To subscribe on data change/resync events
}

type Deps struct {
	InfraDeps localdeps.PluginInfraDeps
}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() error {
	// Initialize plugin fields
	plugin.exampleConfigurator = &ExampleConfigurator{plugin.InfraDeps.ServiceLabel}
	plugin.resyncChannel = make(chan datasync.ResyncEvent)
	plugin.changeChannel = make(chan datasync.ChangeEvent)
	plugin.context = context.Background()

	// Start the consumer (ETCD watcher) before the custom plugin configurator is initialized
	go plugin.consumer()

	// Now initialize the plugin configurator
	plugin.exampleConfigurator.Init()

	go func() {
		// Show simple ETCD CRUD
		plugin.etcdPublisher()
		// Show transactions
		plugin.etcdTxnPublisher()
	}()

	// Subscribe watcher to be able to watch on data changes and resync events
	plugin.subscribeWatcher()

	log.Info("Initialization of the custom plugin for the ETCD example is completed")

	return nil
}

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime
func (plugin *ExamplePlugin) Close() error {
	plugin.exampleConfigurator.Close()
	plugin.watchDataReg.Close()
	return nil
}

/*************************
 * Example plugin config *
 *************************/

// ExampleConfigurator usually initializes configuration-specific fields or other tasks (e.g. defines GOVPP channels
// if they are used, checks VPP message compatibility etc.)
type ExampleConfigurator struct {
	ServiceLabel servicelabel.ReaderAPI
}

// Init members of configurator
func (configurator *ExampleConfigurator) Init() (err error) {
	// There is nothing to init in the example
	log.Info("Custom plugin configurator initialized")

	// Now the configurator is initialized and the watcher is already running (started in plugin initialization),
	// so publisher is used to put data to ETCD


	return err
}

// Close function for example plugin (just for representation, there is nothing to close in the example)
func (configurator *ExampleConfigurator) Close() {}

/*************
 * ETCD call *
 *************/

const etcdIndex string = "index"

// KeyProtoValWriter creates a simple data, then demonstrates CRUD operations with ETCD
func (plugin *ExamplePlugin) etcdPublisher() {
	time.Sleep(3 * time.Second)

	// Convert data to the generated proto format
	exampleData := plugin.buildData("string1", 0, true)

	// PUT: examplePut demonstrates how to use the Data Broker Put() API to create (or update) a simple data
	// structure into ETCD
	plugin.Publisher.Put(etcdKeyPrefixLabel(plugin.InfraDeps.ServiceLabel.GetAgentLabel(), etcdIndex), exampleData)

	// Prepare different set of data
	exampleData = plugin.buildData("string2", 1, false)

	// UPDATE: Put() performs both create operations (if index does not exist) and update operations
	// (if the index exists)
	plugin.Publisher.Put(etcdKeyPrefixLabel(plugin.InfraDeps.ServiceLabel.GetAgentLabel(), etcdIndex), exampleData)

	//todo // GET: exampleGet demonstrates how to use the Data Broker Get() API to read a simple data structure from ETCD
	//result := etcd_example.EtcdExample{}
	//found, _, err := plugin.Publisher.GetValue(etcdKeyPrefixLabel(plugin.ServiceLabel.GetAgentLabel(), etcdIndex), &result)
	//if err != nil {
	//	log.Error(err)
	//}
	//if found {
	//	log.Infof("Data read from ETCD data store. Values: %v, %v, %v",
	//		result.StringVal, result.Uint32Val, result.BoolVal)
	//} else {
	//	log.Error("Data not found")
	//}

	// DELETE: demonstrates how to use the Data Broker Delete() API to delete a simple data structure from ETCD
	//todo plugin.Publisher.Delete(etcdKeyPrefixLabel(plugin.ServiceLabel.GetAgentLabel(), etcdIndex))
}

// KeyProtoValWriter creates a simple data, then demonstrates transaction operations with ETCD
func (plugin *ExamplePlugin) etcdTxnPublisher() {
	log.Info("Preparing bridge domain data")
	// Get data broker to communicate with ETCD
	cfg := &etcdv3.Config{}

	configFile := os.Getenv("ETCDV3_CONFIG")
	if configFile != "" {
		err := config.ParseConfigFromYamlFile(configFile, cfg)
		if err != nil {
			log.Fatal(err)
		}
	}
	etcdConfig, err := etcdv3.ConfigToClientv3(cfg)
	if err != nil {
		log.Fatal(err)
	}

	bDB, _ := etcdv3.NewEtcdConnectionWithBytes(*etcdConfig, log)
	publisher := kvproto.NewProtoWrapperWithSerializer(bDB, &keyval.SerializerJSON{}).
		NewBroker(plugin.InfraDeps.ServiceLabel.GetAgentPrefix())

	time.Sleep(3 * time.Second)

	// This is how to use the Data Broker Txn API to create a new transaction. It is called from the HTTP handler
	// when a user triggers the creation of a new transaction via REST
	putTxn := publisher.NewTxn()
	for i := 1; i <= 3; i++ {
		exampleData1 := plugin.buildData("string", uint32(i), true)
		// putTxn.Put demonstrates how to use the Data Broker Txn Put() API. It is called from the HTTP handler
		// when a user invokes the REST API to add a new Put() operation to the transaction
		putTxn = putTxn.Put(etcdKeyPrefixLabel(plugin.InfraDeps.ServiceLabel.GetAgentLabel(), etcdIndex+strconv.Itoa(i)), exampleData1)
	}
	// putTxn.Commit() demonstrates how to use the Data Broker Txn Commit() API. It is called from the HTTP handler
	// when a user invokes the REST API to commit a transaction.
	err = putTxn.Commit()
	if err != nil {
		log.Error(err)
	}
	// Another transaction chain to demonstrate delete operations. Put and Delete operations can be used together
	// within one transaction
	deleteTxn := publisher.NewTxn()
	for i := 1; i <= 3; i++ {
		// deleteTxn.Delete demonstrates how to use the Data Broker Txn Delete() API. It is called from the
		// HTTP handler when a user invokes the REST API to add a new Delete() operation to the transaction.
		// Put and Delete operations can be combined in the same transaction chain
		deleteTxn = deleteTxn.Delete(etcdKeyPrefixLabel(plugin.InfraDeps.ServiceLabel.GetAgentLabel(), etcdIndex+strconv.Itoa(i)))
	}

	// Commit transactions to data store. Transaction executes multiple operations in a more efficient way in
	// contrast to executing them one by one.
	err = deleteTxn.Commit()
	if err != nil {
		log.Error(err)
	}
}

// The ETCD key prefix used for this example
func etcdKeyPrefix(agentLabel string) string {
	return "/vnf-agent/" + agentLabel + "/api/v1/example/db/simple/"
}

// The ETCD key (the key prefix + label)
func etcdKeyPrefixLabel(agentLabel string, index string) string {
	return etcdKeyPrefix(agentLabel) + index
}

/***********
 * KeyValProtoWatcher *
 ***********/

// Consumer (watcher) is subscribed to watch on data store changes. Change arrives via data change channel and
// its key is parsed
func (plugin *ExamplePlugin) consumer() {
	log.Print("KeyValProtoWatcher started")
	for {
		select {
		case dataChng := <-plugin.changeChannel:
			// If event arrives, the key is extracted and used together with the expected prefix to
			// identify item
			key := dataChng.GetKey()
			if strings.HasPrefix(key, etcdKeyPrefix(plugin.InfraDeps.ServiceLabel.GetAgentLabel())) {
				var value, previousValue etcd_example.EtcdExample
				// The first return value is diff - boolean flag whether previous value exists or not
				err := dataChng.GetValue(&value)
				if err != nil {
					log.Error(err)
				}
				diff, err := dataChng.GetPrevValue(&previousValue)
				if err != nil {
					log.Error(err)
				}
				log.Infof("Event arrived to etcd eventHandler, key %v, update: %v, change type: %v,",
					dataChng.GetKey(), diff, dataChng.GetChangeType())
			}
			// Another strings.HasPrefix(key, etcd prefix) ...
		case <-plugin.context.Done():
			log.Warnf("Stop watching events")
		}

	}
}

// KeyValProtoWatcher is subscribed to data change channel and resync channel. ETCD watcher adapter is used for this purpose
func (plugin *ExamplePlugin) subscribeWatcher() (err error) {
	plugin.watchDataReg, err = plugin.Watcher.
		Watch("Example etcd plugin", plugin.changeChannel, plugin.resyncChannel, etcdKeyPrefix(plugin.InfraDeps.ServiceLabel.GetAgentLabel()))
	if err != nil {
		return err
	}

	log.Info("KeyValProtoWatcher subscribed")

	return nil
}

// Create simple ETCD data structure with provided data values
func (plugin *ExamplePlugin) buildData(stringVal string, uint32Val uint32, boolVal bool) *etcd_example.EtcdExample {
	return &etcd_example.EtcdExample{
		StringVal: stringVal,
		Uint32Val: uint32Val,
		BoolVal:   boolVal,
	}
}
