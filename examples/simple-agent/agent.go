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

package main

import (
	"log"

	"go.ligato.io/cn-infra/v2/agent"
	"go.ligato.io/cn-infra/v2/datasync/resync"
	"go.ligato.io/cn-infra/v2/db/keyval/etcd"
	"go.ligato.io/cn-infra/v2/db/keyval/redis"
	"go.ligato.io/cn-infra/v2/db/sql/cassandra"
)

func main() {

	// Create agent with connector plugins
	a := agent.NewAgent(agent.AllPlugins(
		&etcd.DefaultPlugin,
		&redis.DefaultPlugin,
		&cassandra.DefaultPlugin,
		&resync.DefaultPlugin,
	))

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
