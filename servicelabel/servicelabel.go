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
	"github.com/namsral/flag"
)

// default service key prefix, can be changed in the build time using ldflgs, e.g.
// -ldflags '-X github.com/ligato/cn-infra/servicelabel.agentPrefix=/xyz/'
var agentPrefix = "/vnf-agent/"

type Plugin struct {
	// MicroserviceLabel identifies particular VNF.
	// Used primarily as a key prefix to ETCD data store.
	MicroserviceLabel string
}

var microserviceLabelFlag string

func init() {
	flag.StringVar(&microserviceLabelFlag, "microservice-label", "vpp1", 		"microservice label; also set via 'MICROSERVICE_LABEL' env variable.")
}

// Init is called at plugin initialization.
func (p *Plugin) Init() error {
	if p.MicroserviceLabel == "" {
		p.MicroserviceLabel = microserviceLabelFlag
	}
	return nil
}

// Close is called at plugin cleanup phase.
func (p *Plugin) Close() error {
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
	return agentPrefix + microserviceLabel + "/"
}

// GetAllAgentsPrefix returns the string that is supposed to be used as the prefix for configuration
// subtree of the particular VPP Agent instance (e.g. in ETCD).
func (p *Plugin) GetAllAgentsPrefix() string {
	return agentPrefix
}
