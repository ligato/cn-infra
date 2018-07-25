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
	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/db/cryptodata"
	"io/ioutil"
	"encoding/pem"
	"crypto/x509"
	"crypto/rsa"
	"github.com/pkg/errors"
	"encoding/base64"
	"github.com/ligato/cn-infra/db/keyval"
	"reflect"
	"github.com/ligato/cn-infra/examples/cryptodata-proto-plugin/ipsec"
)

// PluginName represents name of plugin.
const PluginName = "example"

// TunnelName represents name of test tunnel
const TunnelName = "tunnel"

func main() {
	// Start Agent with ExamplePlugin using ETCDPlugin, CryptoDataPlugin, logger and service label.
	p := &ExamplePlugin{
		Deps: Deps{
			Log:          logging.ForPlugin(PluginName),
			ServiceLabel: &servicelabel.DefaultPlugin,
			KvProto:      &etcd.DefaultPlugin,
			CryptoData:   &cryptodata.DefaultPlugin,
		},
		exampleFinished: make(chan struct{}),
	}

	if err := agent.NewAgent(
		agent.AllPlugins(p),
		agent.QuitOnClose(p.exampleFinished),
	).Run(); err != nil {
		log.Fatal(err)
	}
}

// Deps lists dependencies of ExamplePlugin.
type Deps struct {
	Log          logging.PluginLogger
	ServiceLabel servicelabel.ReaderAPI
	KvProto      keyval.KvProtoPlugin
	CryptoData   cryptodata.ClientAPI
}

// ExamplePlugin demonstrates the usage of cryptodata API.
type ExamplePlugin struct {
	Deps
	exampleFinished chan struct{}
}

// String returns plugin name
func (plugin *ExamplePlugin) String() string {
	return PluginName
}

// Init starts the consumer.
func (plugin *ExamplePlugin) Init() error {
	// Read public key
	publicKey, err := readPublicKey("../cryptodata-lib/key-pub.pem")
	if err != nil {
		return err
	}

	// Prepare data
	ip1, err := plugin.encryptData("192.168.0.1", publicKey)
	if err != nil {
		return err
	}
	ip2, err := plugin.encryptData("192.168.0.3", publicKey)
	if err != nil {
		return err
	}
	encryptedData := &ipsec.TunnelInterfaces_Tunnel{
		Name: TunnelName,
		IpAddresses: []string{
			ip1,
			ip2,
		},
	}
	plugin.Log.Infof("Putting value %v", encryptedData)

	// Prepare path for storing the data
	key := plugin.etcdKey(ipsec.TunnelKey(TunnelName))

	// Prepare broker
	broker := plugin.KvProto.NewBroker(keyval.Root)

	// Put proto data to ETCD
	err = broker.Put(key, encryptedData)
	if err != nil {
		return err
	}

	// Wrap broker with crypto layer
	brokerWrapped := cryptodata.ProtoBrokerWrapper{
		ProtoBroker: broker,
		CryptoMap: map[reflect.Type][][]string{
			reflect.TypeOf(&ipsec.TunnelInterfaces_Tunnel{}): {{"IpAddresses"}},
		},
		DecryptFunc: plugin.CryptoData.DecryptData,
	}

	// Get proto data from ETCD and decrypt them with crypto layer
	decryptedData := &ipsec.TunnelInterfaces_Tunnel{}
	_, _, err = brokerWrapped.GetValue(key, decryptedData)
	if err != nil {
		return err
	}
	plugin.Log.Infof("Got value %v", decryptedData)

	// Close agent and example
	close(plugin.exampleFinished)

	return nil
}

// Close closes ExamplePlugin
func (plugin *ExamplePlugin) Close() error {
	return nil
}

// The ETCD key prefix used for this example
func (plugin *ExamplePlugin) etcdKey(label string) string {
	return "/vnf-agent/" + plugin.ServiceLabel.GetAgentLabel() + "/api/v1/example/db/simple/" + label
}

// encryptData first encrypts the provided value using crypto layer and then encodes
// the data with base64 for JSON compatibility
func (plugin *ExamplePlugin) encryptData(value string, publicKey *rsa.PublicKey) (string, error) {
	encryptedValue, err := plugin.CryptoData.EncryptData([]byte(value), publicKey)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(encryptedValue), nil
}

// readPublicKey reads rsa public key from PEM file on provided path
func readPublicKey(path string) (*rsa.PublicKey, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM for key " + path)
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	publicKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("failed to convert public key to rsa.PublicKey")
	}

	return publicKey, nil
}
