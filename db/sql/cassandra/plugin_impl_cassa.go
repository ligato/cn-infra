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

package cassandra

import (
	"github.com/gocql/gocql"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/db/sql"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/willfaught/gockle"
)

// HealthCheck is a structure used to represent table to help verify health status
type HealthCheck struct {
	ID           gocql.UUID `cql:"id" pk:"id"`
	HealthStatus bool       `cql:"healthStatus"`
}

// Plugin implements Plugin interface therefore can be loaded with other plugins
type Plugin struct {
	Deps // inject

	clientConfig *ClientConfig
	session      gockle.Session
}

// Deps is here to group injected dependencies of plugin
// to not mix with other plugin fields.
type Deps struct {
	local.PluginInfraDeps // inject
}

// Init is called at plugin startup. The session to etcd is established.
func (p *Plugin) Init() (err error) {
	if p.session != nil {
		return nil // skip initialization
	}

	// Retrieve config
	var cfg Config
	found, err := p.PluginConfig.GetValue(&cfg)
	// need to be strict about config presence for ETCD
	if !found {
		p.Log.Info("cassandra client config not found ", p.PluginConfig.GetConfigName(),
			" - skip loading this plugin")
		return nil
	}
	if err != nil {
		return err
	}

	// Init session
	p.clientConfig, err = ConfigToClientConfig(&cfg)
	if err != nil {
		return err
	}

	return nil
}

// AfterInit is called by the Agent Core after all plugins have been initialized.
func (p *Plugin) AfterInit() error {
	if p.session == nil && p.clientConfig != nil {
		session, err := CreateSessionFromConfig(p.clientConfig)
		if err != nil {
			return err
		}

		p.session = gockle.NewSession(session)
	}

	// Register for providing status reports (polling mode)
	if p.StatusCheck != nil && p.session != nil {
		p.StatusCheck.Register(core.PluginName(p.String()), func() (statuscheck.PluginState, error) {
			broker := p.NewBroker()
			err := createKeyspace(p, broker)
			if err == nil {
				err = createTable(p, broker)
				if err == nil {
					err = insertHealthCheckItem(p, broker)
					if err == nil {
						err = getHealthCheckItem(p, broker)
						if err == nil {
							return statuscheck.OK, nil
						}
					}
				}
			}
			return statuscheck.Error, err
		})
	} else {
		p.Log.Warnf("Unable to start status check for Cassandra")
	}

	return nil
}

// FromExistingSession is used mainly for testing
func FromExistingSession(session gockle.Session) *Plugin {
	return &Plugin{session: session}
}

// NewBroker returns a Broker instance to work with Cassandra Data Base
func (p *Plugin) NewBroker() sql.Broker {
	return NewBrokerUsingSession(p.session)
}

// Close resources
func (p *Plugin) Close() error {
	p.session.Close()
	return nil
}

// String returns if set Deps.PluginName or "cassa-client" otherwise
func (p *Plugin) String() string {
	if len(p.Deps.PluginName) == 0 {
		return "cassa-client"
	}
	return string(p.Deps.PluginName)
}

// SchemaName returns the schema name for HealthCheck table
func (entity *HealthCheck) SchemaName() string {
	return "healthStatusCheck"
}

// createKeyspace used to create keyspace for health check to verify health status
func createKeyspace(p *Plugin, broker sql.Broker) (err error) {
	err = broker.Exec(`CREATE KEYSPACE IF NOT EXISTS healthStatusCheck with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };`)
	if err != nil {
		p.Log.Errorf("Unable to create keyspace in Cassandra")
		return err
	}
	return nil
}

// createTable used to create table for health check to verify health status
func createTable(p *Plugin, broker sql.Broker) (err error) {
	err = broker.Exec(`CREATE TABLE IF NOT EXISTS healthStatusCheck.healthCheck (
		id uuid PRIMARY KEY,
		healthStatus boolean
	);`)
	if err != nil {
		p.Log.Errorf("Unable to create table in Cassandra")
		return err
	}
	return nil
}

// insertHealthCheckItem used to insert an item in the health check table to verify health status
func insertHealthCheckItem(p *Plugin, broker sql.Broker) (err error) {
	healthCheckItem := &HealthCheck{HealthStatus: true}
	err = broker.Put(sql.Exp("id=c37d661d-7e61-49ea-96a5-68c34e83db3a"), healthCheckItem)
	if err != nil {
		p.Log.Errorf("Unable to insert data in Cassandra")
		return err
	}
	return nil
}

// getHealthCheckItem used to retrieve an item from the health check table to verify health status
func getHealthCheckItem(p *Plugin, broker sql.Broker) (err error) {
	healthCheckTable := &HealthCheck{}
	healthStatus := &[]HealthCheck{}
	err = sql.SliceIt(healthStatus, broker.ListValues(sql.FROM(healthCheckTable,
		sql.WHERE(sql.Field(&healthCheckTable.ID, sql.EQ("c37d661d-7e61-49ea-96a5-68c34e83db3a"))))))
	if err != nil {
		p.Log.Errorf("Unable to retrieve data from Cassandra")
		return err
	}

	if len(*healthStatus) <= 0 && !(*healthStatus)[0].HealthStatus {
		p.Log.Errorf("Record not found in Cassandra")
		return err
	}

	return nil
}
