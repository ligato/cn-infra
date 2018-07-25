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
	"github.com/ligato/cn-infra/config"
	"io/ioutil"
	"encoding/pem"
	"crypto/x509"
	"crypto/rsa"
	"github.com/pkg/errors"
	"encoding/base64"
	"fmt"
)

// PluginName represents name of plugin.
const PluginName = "example"

func main() {
	// Start Agent with ExamplePlugin using ETCDPlugin CryptoDataPlugin, logger and service label.
	p := &ExamplePlugin{
		Deps: Deps{
			Log:          logging.ForPlugin(PluginName),
			ServiceLabel: &servicelabel.DefaultPlugin,
			CryptoData:   &cryptodata.DefaultPlugin,
		},
	}

	if err := agent.NewAgent(agent.AllPlugins(p)).Run(); err != nil {
		log.Fatal(err)
	}
}

// Deps lists dependencies of ExamplePlugin.
type Deps struct {
	Log          logging.PluginLogger
	ServiceLabel servicelabel.ReaderAPI
	CryptoData   cryptodata.Client
}

// ExamplePlugin demonstrates the usage of datasync API.
type ExamplePlugin struct {
	Deps
}

// String returns plugin name
func (plugin *ExamplePlugin) String() string {
	return PluginName
}

// Init starts the consumer.
func (plugin *ExamplePlugin) Init() error {
	// Read public key
	bytes, err := ioutil.ReadFile("../cryptodata-lib/key-pub.pem")
	if err != nil {
		return err
	}
	block, _ := pem.Decode(bytes)
	if block == nil {
		return errors.New("failed to decode PEM for key key-pub.pem")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	publicKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return errors.New("failed to convert public key to rsa.PublicKey")
	}

	// Create ETCD connection
	etcdFileConfig := &etcd.Config{}
	err = config.ParseConfigFromYamlFile("etcd.conf", etcdFileConfig)
	if err != nil {
		return err
	}

	etcdConfig, err := etcd.ConfigToClient(etcdFileConfig)
	if err != nil {
		return err
	}

	db, err := etcd.NewEtcdConnectionWithBytes(*etcdConfig, plugin.Log)
	if err != nil {
		return err
	}

	// Wrap ETCD connection with crypto layer
	dbWrapped := plugin.CryptoData.Wrap(db, cryptodata.NewDecrypterJSON())

	// Prepare data
	value := "hello-world"
	encryptedValue, err := plugin.CryptoData.EncryptData([]byte(value), publicKey)
	if err != nil {
		return err
	}
	encryptedBase64Value := base64.URLEncoding.EncodeToString(encryptedValue)
	encryptedJSON := fmt.Sprintf(`{"encrypted":true,"value":{"payload":"$crypto$%v"}}`, encryptedBase64Value)
	plugin.Log.Infof("Putting value %v", encryptedJSON)

	// Put and get
	key := plugin.etcdKey("value")
	err = dbWrapped.Put(key, []byte(encryptedJSON))
	if err != nil {
		return err
	}

	decryptedJSON, _, _, err := dbWrapped.GetValue(key)
	if err != nil {
		return err
	}

	plugin.Log.Infof("Got value %v", string(decryptedJSON))
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
