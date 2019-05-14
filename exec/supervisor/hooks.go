// Copyright (c) 2019 Cisco and/or its affiliates.
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

package supervisor

import (
	"os/exec"
)

func (p *Plugin) watchEvents() {
	for {
		processInfo, ok :=<-p.hookChan
		if !ok {
			return
		}

		// find and execute related hooks
		for _, hook := range p.config.Hooks {
			if hook.ProgramName == processInfo.name && hook.EventType == string(processInfo.state) {
				p.Log.Debugf("executing hook for %s, state %s", hook.ProgramName, hook.EventType)
				out, err := exec.Command(hook.Cmd, hook.CmdArgs...).CombinedOutput()
				if err != nil {
					p.Log.Errorf("%v", err)
				}
				if len(out) > 0 {
					p.Log.Debugf("%s", out)
				}
			}
		}
	}
}
