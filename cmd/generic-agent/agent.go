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
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/core/flavours"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/namsral/flag"
	"os"
	"time"
)

func main() {
	logroot.Logger().SetLevel(logging.DebugLevel)

	f := flavours.Generic{}
	f.RegisterFlags()
	flag.Parse()

	err := f.ApplyConfig()
	if err != nil {
		logroot.Logger().Error(err)
		os.Exit(1)
	}
	f.Inject()

	agent := core.NewAgent(logroot.Logger(), 15*time.Second, f.Plugins()...)
	core.EventLoopWithInterrupt(agent, nil)
}
