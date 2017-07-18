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

package logrus

import (
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/plugin"
)

// Plugin implements logging plugin using Logrus library.
type Plugin struct {
	*plugin.Skeleton
}

// NewLogrusPlugin creates a new instance of Logrus logging plugin.
func NewLogrusPlugin() *Plugin {
	factory := func(name string) (logging.Logger, error) {
		l, err := NewNamed(name)
		if err != nil {
			return l, err
		}
		l.SetLevel(logging.DebugLevel)
		return l, err
	}
	sk := plugin.NewSkeleton(factory, func() logging.Registry { return LoggerRegistry })
	return &Plugin{sk}
}
