// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"time"

	"os"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/tlsutil"
	"github.com/ghodss/yaml"
)

type yamlConfig struct {
	Endpoints             []string      `json:"endpoints"`
	DialTimeout           time.Duration `json:"dial-timeout"`
	InsecureTransport     bool          `json:"insecure-transport"`
	InsecureSkipTLSVerify bool          `json:"insecure-skip-tls-verify"`
	Certfile              string        `json:"cert-file"`
	Keyfile               string        `json:"key-file"`
	CAfile                string        `json:"ca-file"`
}

// default timeout for connecting to etcd.
const defaultTimeout = 1 * time.Second

// configFromFile loads the Etcd client configuration from the
// specified file. If the specified file is valid and contains
// valid configuration, the parsed client configuration is
// returned; otherwise, an error is returned.
func configFromFile(fpath string) (*clientv3.Config, error) {
	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	yc := &yamlConfig{}

	err = yaml.Unmarshal(b, yc)
	if err != nil {
		return nil, err
	}

	timeout := defaultTimeout

	if yc.DialTimeout != 0 {
		timeout = yc.DialTimeout
	}

	cfg := &clientv3.Config{
		Endpoints:   yc.Endpoints,
		DialTimeout: timeout,
	}

	if yc.InsecureTransport {
		cfg.TLS = nil
		return cfg, nil
	}

	var (
		cert *tls.Certificate
		cp   *x509.CertPool
	)

	if yc.Certfile != "" && yc.Keyfile != "" {
		cert, err = tlsutil.NewCert(yc.Certfile, yc.Keyfile, nil)
		if err != nil {
			return nil, err
		}
	}

	if yc.CAfile != "" {
		cp, err = tlsutil.NewCertPool([]string{yc.CAfile})
		if err != nil {
			return nil, err
		}
	}

	tlscfg := &tls.Config{
		MinVersion:         tls.VersionTLS10,
		InsecureSkipVerify: yc.InsecureSkipTLSVerify,
		RootCAs:            cp,
	}
	if cert != nil {
		tlscfg.Certificates = []tls.Certificate{*cert}
	}
	cfg.TLS = tlscfg

	return cfg, nil
}

// initRemoteClient initializes the Connection to ETCD. ETCD clientv3
// config file contains the settings of the ETCD connection. A new clientv3
// is created in the function and returned to the caller.
func initRemoteClient(configFile string) (*clientv3.Client, error) {
	var config *clientv3.Config
	var err error

	if configFile != "" {
		config, err = configFromFile(configFile)
		if err != nil {
			return nil, err
		}
	} else if ep := os.Getenv("ETCDV3_ENDPOINTS"); ep != "" {
		config = &clientv3.Config{
			Endpoints:   strings.Split(ep, ","),
			DialTimeout: defaultTimeout}
	} else {
		config = &clientv3.Config{
			Endpoints:   []string{"127.0.0.1:2379"},
			DialTimeout: defaultTimeout}
	}

	var etcdClient *clientv3.Client
	etcdClient, err = clientv3.New(*config)
	if err != nil {
		log.Errorf("Failed to connect to Etcd etcd(s) %v, Error: '%s'", config.Endpoints, err)
		return nil, err
	}
	return etcdClient, nil
}
