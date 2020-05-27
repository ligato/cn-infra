//  Copyright (c) 2020 Cisco and/or its affiliates.
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
	"go.ligato.io/cn-infra/v2/datasync/kvdbsync"
	"go.ligato.io/cn-infra/v2/datasync/resync"
	"go.ligato.io/cn-infra/v2/db/keyval"
	"go.ligato.io/cn-infra/v2/db/keyval/etcd"
	etcdexample "go.ligato.io/cn-infra/v2/examples/model"
	"go.ligato.io/cn-infra/v2/health/statuscheck"
	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/servicelabel"
)

func main() {
	app := infra.NewApp(
		infra.Provide(
			provideStatusCheck,
			provideResync,
			provideEtcd,
			provideKVDBSync,
			provideServiceLabel,
			provideLogger,
		),
		infra.Invoke(Example),
	)
	app.Run()
}

func Example(publisher *kvdbsync.Plugin, svclabel servicelabel.ReaderAPI, log logging.PluginLogger) {
	log.Print("KeyValPublisher started")

	// Convert data into the proto format.
	exampleData := &etcdexample.EtcdExample{
		StringVal: "bla",
		Uint32Val: 3,
		BoolVal:   true,
	}

	// PUT: demonstrate how to use the Data Broker Put() API to store
	// a simple data structure into ETCD.
	label := "/infra/" + svclabel.GetAgentLabel() + "data1"

	log.Infof("Write data to %v", label)

	if err := publisher.Put(label, exampleData); err != nil {
		log.Fatal(err)
	}
}

func provideServiceLabel() servicelabel.ReaderAPI {
	//p := &servicelabel.Plugin{}
	//p.SetName("servicelabel")
	//return p
	return &servicelabel.DefaultPlugin
}

func provideLogger() logging.PluginLogger {
	return logging.NewParentLogger("infra", logging.DefaultRegistry)
}

func provideStatusCheck(log logging.PluginLogger) statuscheck.PluginStatusWriter {
	p := &statuscheck.Plugin{}
	p.SetName("statuscheck")
	p.Log = log
	return p
}

func provideResync() resync.Subscriber {
	p := &resync.Plugin{}
	p.SetName("resync")
	p.Log = logging.NewParentLogger("resync2", logging.DefaultRegistry)
	return p
}

func provideEtcd(statuscheck statuscheck.PluginStatusWriter) keyval.KvProtoPlugin {
	p := &etcd.Plugin{}
	p.SetName("statuscheck")
	p.StatusCheck = statuscheck
	p.Log = logging.NewParentLogger("statuscheck", logging.DefaultRegistry)
	return p
}

func provideKVDBSync(kvdb keyval.KvProtoPlugin, resyncSub resync.Subscriber) *kvdbsync.Plugin {
	p := &kvdbsync.Plugin{}
	p.SetName("kvdbsync")
	p.KvPlugin = kvdb
	p.ResyncOrch = resyncSub
	p.ServiceLabel = &servicelabel.DefaultPlugin
	p.Log = logging.NewParentLogger("kvdbsync", logging.DefaultRegistry)
	return p
}
