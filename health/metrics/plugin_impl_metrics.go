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

// Package metrics implements Prometheus health/metrics handlers.
package metrics

import (
	"net/http"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/logging"
	log "github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/namsral/flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/unrolled/render"
)

const (
	defaultPluginName  string = "HEALTH-METRICS"
	DefaultMetricsPath string = "/metrics" // default Prometheus metrics URL
	agentName          string = "agent"
	namespace          string = ""
	subsystem          string = ""
	serviceLabel       string = "service"
	dependencyLabel    string = "dependency"
	buildVersionLabel  string = "build_version"
	buildDateLabel     string = "build_date"
	serviceHealthName  string = "service_health"
	// Adapt Ligato status code for now.
	// TODO: Consolidate with that from the "Common Container Telemetry" proposal.
	//serviceHealthHelp    string = "The health of the serviceLabel 0 = INIT, 1 = UP, 2 = DOWN, 3 = OUTAGE"
	serviceHealthHelp    string = "The health of the serviceLabel 0 = INIT, 1 = OK, 2 = ERROR"
	dependencyHealthName string = "service_dependency_health"
	// Adapt Ligato status code for now.
	// TODO: Consolidate with that from the "Common Container Telemetry" proposal.
	//dependencyHealthHelp string = "The health of the dependencyLabel 0 = INIT, 1 = UP, 2 = DOWN, 3 = OUTAGE"
	dependencyHealthHelp string = "The health of the dependencyLabel 0 = INIT, 1 = OK, 2 = ERROR"
	serviceInfoName      string = "service_info"
	serviceInfoHelp      string = "Build info for the service.  Value is always 1, build info is in the tags."
)

var (
	httpPort string
)

// init is here only for parsing program arguments
func init() {
	flag.StringVar(&httpPort, "prometheus-http-port", rest.DefaultHTTPPort,
		"Listening port for the Agent's Prometheus health/metrics port.")
}

// Plugin struct holds all plugin-related data.
type Plugin struct {
	Deps

	customProbe bool
}

// Deps lists dependencies of the Prometheus plugin.
type Deps struct {
	local.PluginLogDeps                               // inject
	HTTP                *rest.Plugin                  // inject (optional)
	StatusCheck         statuscheck.AgentStatusReader // inject
}

// Init may create a new (custom) instance of HTTP if the injected instance uses
// different HTTP port than requested.
func (p *Plugin) Init() (err error) {
	// Start Init() and AfterInit() for new Prometheus in case the port is different
	// from agent http.
	if p.HTTP.HTTPport != httpPort {
		childPlugNameHTTP := p.String() + "-HTTP"
		p.HTTP = &rest.Plugin{
			Deps: rest.Deps{
				Log:        logging.ForPlugin(childPlugNameHTTP, p.Log),
				PluginName: core.PluginName(childPlugNameHTTP),
				HTTPport:   httpPort,
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

	p.RegisterGauge(
		namespace,
		subsystem,
		serviceHealthName,
		serviceHealthHelp,
		prometheus.Labels{serviceLabel: agentName},
		p.getServiceHealth,
	)

	agentStatus := p.StatusCheck.GetAgentStatus()
	p.RegisterGauge(
		namespace,
		subsystem,
		serviceInfoName,
		serviceInfoHelp,
		prometheus.Labels{
			serviceLabel:      agentName,
			buildVersionLabel: agentStatus.BuildVersion,
			buildDateLabel:    agentStatus.BuildDate},
		func() float64 { return 1 },
	)

	return nil
}

// AfterInit registers HTTP handlers.
func (p *Plugin) AfterInit() error {
	if p.HTTP != nil {
		if p.StatusCheck != nil {
			p.Log.Infof("Starting Prometheus metrics handlers on port %v", p.HTTP.HTTPport)
			p.HTTP.RegisterHTTPHandler(DefaultMetricsPath, p.metricsHandler, "GET")
		} else {
			p.Log.Info("Unable to register Prometheus metrics handlers, StatusCheck is nil")
		}
	} else {
		p.Log.Info("Unable to register Prometheus metrics handlers, HTTP is nil")
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

// metricsHandler handles Prometheus metrics collection.
func (p *Plugin) metricsHandler(formatter *render.Render) http.HandlerFunc {
	return promhttp.Handler().ServeHTTP
}

func (p *Plugin) getServiceHealth() float64 {
	agentStatus := p.StatusCheck.GetAgentStatus()
	// Adapt Ligato status code for now.
	// TODO: Consolidate with that from the "Common Container Telemetry" proposal.
	health := float64(agentStatus.State)
	log.StandardLogger().Infof("getServiceHealth(): %f", health)
	return health
}

// RegisterGauge registers custom gauge with specific valueFunc to report status when invoked.
func (p *Plugin) RegisterGauge(namespace string, subsystem string, name string, help string,
	labels prometheus.Labels, valueFunc func() float64) {
	gaugeName := name
	if subsystem != "" {
		gaugeName = subsystem + "_" + gaugeName
	}
	if namespace != "" {
		gaugeName = namespace + "_" + gaugeName
	}
	if err := prometheus.DefaultRegisterer.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			// Namespace, Subsystem, and Name are components of the fully-qualified
			// name of the Metric (created by joining these components with
			// "_"). Only Name is mandatory, the others merely help structuring the
			// name. Note that the fully-qualified name of the metric must be a
			// valid Prometheus metric name.
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      name,

			// Help provides information about this metric. Mandatory!
			//
			// Metrics with the same fully-qualified name must have the same Help
			// string.
			Help: help,

			// ConstLabels are used to attach fixed labels to this metric. Metrics
			// with the same fully-qualified name must have the same label names in
			// their ConstLabels.
			//
			// Note that in most cases, labels have a value that varies during the
			// lifetime of a process. Those labels are usually managed with a metric
			// vector collector (like CounterVec, GaugeVec, UntypedVec). ConstLabels
			// serve only special purposes. One is for the special case where the
			// value of a label does not change during the lifetime of a process,
			// e.g. if the revision of the running binary is put into a
			// label. Another, more advanced purpose is if more than one Collector
			// needs to collect Metrics with the same fully-qualified name. In that
			// case, those Metrics must differ in the values of their
			// ConstLabels. See the Collector examples.
			//
			// If the value of a label never changes (not even between binaries),
			// that label most likely should not be a label at all (but part of the
			// metric name).
			ConstLabels: labels,
		},
		valueFunc,
	)); err == nil {
		log.StandardLogger().Infof("GaugeFunc('%s') registered.", gaugeName)
	} else {
		log.StandardLogger().Errorf("GaugeFunc('%s') registration failed: %s", gaugeName, err)
	}
}

// String returns plugin name if it was injected, defaultPluginName otherwise.
func (p *Plugin) String() string {
	if len(string(p.PluginName)) > 0 {
		return string(p.PluginName)
	}
	return defaultPluginName
}
