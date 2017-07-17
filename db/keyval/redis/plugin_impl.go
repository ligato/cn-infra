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

package redis

import (
	"github.com/ligato/cn-infra/db/keyval/plugin"
	"github.com/ligato/cn-infra/logging"
)

// ProtoPluginRedis implements Plugin interface therefore can be loaded with other plugins
type ProtoPluginRedis struct {
	*plugin.Skeleton
	//TODO `inject:""`	-- Copied from etcdv3/plugin_impl.go.  What should be done here?
}

// NewRedisPlugin creates a new instance of ProtoPluginRedis.
func NewRedisPlugin(pool ConnPool, log logging.Logger) *ProtoPluginRedis {

	skeleton := plugin.NewSkeleton(
		func(log logging.Logger) (plugin.Connection, error) {
			return NewBytesConnectionRedis(pool, log)
		},
	)
	return &ProtoPluginRedis{Skeleton: skeleton}
}
