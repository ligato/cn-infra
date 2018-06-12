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

package servicelabel

import (
	"fmt"

	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/utils/once"
	"github.com/namsral/flag"
)

// Plugin exposes the service label(i.e. the string used to identify the particular VNF) to the other plugins.
type Plugin struct {
	// MicroserviceLabel identifies particular VNF.
	// Used primarily as a key prefix to ETCD data store.
	MicroserviceLabel string
	initOnce          once.ReturnError
}

// OfDifferentAgent sets micorserivce label and returns new instance of Plugin.
// It is meant for watching DB by prefix of different agent. You can pass/inject
// instance of this plugin for example to kvdbsync.
func OfDifferentAgent(microserviceLabel string) *Plugin {
	return &Plugin{MicroserviceLabel: microserviceLabel}
}

var microserviceLabelFlag string

func init() {
	flag.StringVar(&microserviceLabelFlag, "microservice-label", "vpp1", fmt.Sprintf("microservice label; also set via '%v' env variable.", MicroserviceLabelEnvVar))
}

// Init is called at plugin initialization.
func (p *Plugin) Init() error {
	return p.initOnce.Do(p.init)
}

func (p *Plugin) init() error {
	if p.MicroserviceLabel == "" {
		p.MicroserviceLabel = microserviceLabelFlag
	}
	logrus.DefaultLogger().Debugf("Microservice label is set to %v", p.MicroserviceLabel)
	return nil
}

// Close is called at plugin cleanup phase.
func (p *Plugin) close() error {
	// Warning, if you need to actually do anything in close in the future, please
	// a)  Create a field in Plugin:
	//     closeOnce          once.OnceWithError
	// b)  Rename the existing Close() function to close()
	// c)  Create a new Close() function:
	//     func (p *Plugin) Close() error {
	//	     return p.closeOnce.Do(p.close)
	//     }
	return nil
}

// GetAgentLabel returns string that is supposed to be used to distinguish
// (ETCD) key prefixes for particular VNF (particular VPP Agent configuration)
func (p *Plugin) GetAgentLabel() string {
	return p.MicroserviceLabel
}

// GetAgentPrefix returns the string that is supposed to be used as the prefix for configuration of current
// MicroserviceLabel "subtree" of the particular VPP Agent instance (e.g. in ETCD).
func (p *Plugin) GetAgentPrefix() string {
	return agentPrefix + p.MicroserviceLabel + "/"
}

// GetDifferentAgentPrefix returns the string that is supposed to be used as the prefix for configuration
// "subtree" of the particular VPP Agent instance (e.g. in ETCD).
func (p *Plugin) GetDifferentAgentPrefix(microserviceLabel string) string {
	return GetDifferentAgentPrefix(microserviceLabel)
}

// GetAllAgentsPrefix returns the string that is supposed to be used as the prefix for configuration
// subtree of the particular VPP Agent instance (e.g. in ETCD).
func (p *Plugin) GetAllAgentsPrefix() string {
	return GetAllAgentsPrefix()
}
