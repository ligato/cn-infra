// Copyright (c) 2018 Cisco and/or its affiliates.
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

// Package crdexample defines flavor used for the netmesh agent.
package crdexample

import (
	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/examples/crd-plugin/plugins/crd"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/flavors/rpc"
)

const (
	// MicroserviceLabel is the microservice label used by netmesh.
	MicroserviceLabel = "crdexample"

	// KubeConfigAdmin is the default location of kubeconfig with admin credentials.
	KubeConfigAdmin = "/etc/kubernetes/admin.conf"

	// KubeConfigUsage explains the purpose of 'kube-config' flag.
	KubeConfigUsage = "Path to the kubeconfig file to use for the client connection to K8s cluster"
)

// NewAgent returns a new instance of the Agent with plugins.
// It is an alias for core.NewAgent() to implicit use of the FlavorExampleCrd
func NewAgent(opts ...core.Option) *core.Agent {
	return core.NewAgent(&FlavorCrdExample{}, opts...)
}

// WithPlugins for adding custom plugins to your Agent
// <listPlugins> is a callback that uses flavor input to
// inject dependencies for custom plugins that are in output
func WithPlugins(listPlugins func(local *FlavorCrdExample) []*core.NamedPlugin) core.WithPluginsOpt {
	return &withPluginsOpt{listPlugins}
}

// FlavorCrdExample glues together multiple plugins to watch selected k8s
// resources and causes all changes to be reflected in a given store.
type FlavorCrdExample struct {
	// Local flavor is used to access the Infra (logger, service label, status check)
	*local.FlavorLocal
	// RPC flavor for REST-based management.
	*rpc.FlavorRPC
	// Kubernetes State Reflector plugin works as a reflector for policies, pods
	// and namespaces.
	CRD exampleplugincrd.Plugin

	injected bool
}

// Inject sets inter-plugin references.
func (f *FlavorCrdExample) Inject() (allReadyInjected bool) {
	if f.injected {
		return false
	}
	f.injected = true

	if f.FlavorLocal == nil {
		f.FlavorLocal = &local.FlavorLocal{}
	}
	f.FlavorLocal.Inject()
	f.FlavorLocal.ServiceLabel.MicroserviceLabel = MicroserviceLabel
	if f.FlavorRPC == nil {
		f.FlavorRPC = &rpc.FlavorRPC{FlavorLocal: f.FlavorLocal}
	}
	f.FlavorRPC.Inject()

	f.CRD.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("examplecrd")
	f.CRD.Deps.KubeConfig = config.ForPlugin("kube", KubeConfigAdmin, KubeConfigUsage)

	return true
}

// Plugins combines all plugins in the flavor into a slice.
func (f *FlavorCrdExample) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}

// withPluginsOpt is return value of vppLocal.WithPlugins() utility
// to easily define new plugins for the agent based on FlavorExampleCrd
type withPluginsOpt struct {
	callback func(local *FlavorCrdExample) []*core.NamedPlugin
}

// OptionMarkerCore is just for marking implementation that it implements this interface
func (opt *withPluginsOpt) OptionMarkerCore() {}

// Plugins methods is here to implement core.WithPluginsOpt go interface
// <flavor> is a callback that uses flavor input for dependency injection
// for custom plugins (returned as NamedPlugin)
func (opt *withPluginsOpt) Plugins(flavors ...core.Flavor) []*core.NamedPlugin {
	for _, flavor := range flavors {
		if f, ok := flavor.(*FlavorCrdExample); ok {
			return opt.callback(f)
		}
	}

	panic("wrong usage of netmesh.WithPlugin() for other than FlavorExampleCrd")
}
