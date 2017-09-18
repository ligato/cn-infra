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

// Package probe implements the K8s readiness and liveliness probe handlers.
package probe

import (
	"encoding/json"
	"net/http"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/health/statuscheck/model/status"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/namsral/flag"
	"github.com/unrolled/render"
)

const (
	// Default port used for http and probing.
	defaultPort               = "9191"
	livenessProbePath  string = "/liveness"  // liveness probe URL
	readinessProbePath string = "/readiness" // readiness probe URL
)

var (
	httpPort string
)

// init is here only for parsing program arguments
func init() {
	flag.StringVar(&httpPort, "http-probe-port", defaultPort,
		"Listen port for the Agent's HTTPProbe server.")
}

// Plugin struct holds all plugin-related data.
type Plugin struct {
	Deps

	customProbe bool
}

// Deps lists dependencies of the probe plugin.
type Deps struct {
	local.PluginLogDeps                               // inject
	HTTP                *rest.Plugin                  // inject (optional)
	StatusCheck         statuscheck.AgentStatusReader // inject
}

// Init may create a new (custom) instance of HTTP if the injected instance uses
// different HTTP port than requested.
func (p *Plugin) Init() (err error) {
	// Start Init() and AfterInit() for new probe in case the port is different
	// from agent http.
	if p.HTTP.HTTPport != httpPort {
		childPlugNameHTTP := p.String() + "-HTTP"
		p.HTTP = &rest.Plugin{
			Deps: rest.Deps{
				Log:        logging.ForPlugin(childPlugNameHTTP, p.Log),
				PluginName: core.PluginName(childPlugNameHTTP),
				HTTPport: httpPort,
			},
		}
		err := p.HTTP.Init()
		if err != nil {
			return err
		}
		err = p.HTTP.AfterInit()
		if err != nil {
			return err
		}

		p.customProbe = true
	}
	return nil
}

// AfterInit registers HTTP handlers for liveness and readiness probes.
func (p *Plugin) AfterInit() error {
	if p.HTTP != nil {
		if p.StatusCheck != nil {
			p.Log.Infof("Starting health http-probe on port %v", p.HTTP.HTTPport)
			p.HTTP.RegisterHTTPHandler(livenessProbePath, p.livenessProbeHandler, "GET")
			p.HTTP.RegisterHTTPHandler(readinessProbePath, p.readinessProbeHandler, "GET")

		} else {
			p.Log.Info("Unable to register http-probe handler, StatusCheck is nil")
		}
	} else {
		p.Log.Info("Unable to register http-probe handler, HTTP is nil")
	}

	return nil
}

// Close shutdowns HTTP if a custom instance was created in Init().
func (p *Plugin) Close() error {
	if p.customProbe {
		_, err := safeclose.CloseAll(p.HTTP)
		return err
	}

	return nil
}

// readinessProbeHandler handles k8s readiness probe.
func (p *Plugin) readinessProbeHandler(formatter *render.Render) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		stat := p.StatusCheck.GetAgentStatus()
		statJSON, _ := json.Marshal(stat)
		if stat.State == status.OperationalState_OK {
			w.WriteHeader(http.StatusOK)
			w.Write(statJSON)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(statJSON)
		}
	}
}

// livenessProbeHandler handles k8s liveness probe.
func (p *Plugin) livenessProbeHandler(formatter *render.Render) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		stat := p.StatusCheck.GetAgentStatus()
		statJSON, _ := json.Marshal(p.StatusCheck.GetAgentStatus())

		if stat.State == status.OperationalState_INIT || stat.State == status.OperationalState_OK {
			w.WriteHeader(http.StatusOK)
			w.Write(statJSON)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(statJSON)
		}
	}
}

// String returns plugin name if it was injected, "HEALTH_RPC_PROBES" otherwise.
func (p *Plugin) String() string {
	if len(string(p.PluginName)) > 0 {
		return string(p.PluginName)
	}
	return "HEALTH_RPC_PROBES"
}
