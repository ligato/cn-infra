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

package grpc

import (
	"github.com/go-errors/errors"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/cn-infra/wiring"
	"github.com/ligato/cn-infra/core"
)

// WithRest wires up the HTTP dependency for Plugin
// If overwrite is false, existing values will not be overwritten
// If handlers is provided, that will be configured as the HTTP handler, otherwise a default will be used
func WithRest(overwrite bool, handlers...rest.HTTPHandlers) wiring.Wiring {
	ret := func (plugin core.Plugin) error {
		p, ok := plugin.(*Plugin)
		if ok {
			if overwrite || p.HTTP == nil {
				if len(handlers) > 0 {
					p.HTTP = handlers[0]
				} else {
					rp := &rest.Plugin{}
					httpWiring := wiring.ComposeWirings(
						rest.DefaultWiring(false),
						rest.WithLogFactory(true,p.Log))
					if err := rp.Wire(httpWiring); err != nil {
						return err
					}
					p.HTTP = rp
				}
			}
			return nil
		}
		return errors.Errorf("Could not convert core.Plugin to *%s.Plugin",packageName)
	}
	return ret
}
