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

package adapters

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/utils/safeclose"
)

// TransportAggregator is cumulative adapter which contains all available transport types
type TransportAggregator struct {
	Adapters []datasync.TransportAdapter
}

// AggregatedRegistration is cumulative adapter which contains all available transport types
type AggregatedRegistration struct {
	Registrations []datasync.WatchDataRegistration
}

// WatchData subscribes to every transport available within transport aggregator
func (ta *TransportAggregator) WatchData(resyncName string, changeChan chan datasync.ChangeEvent, resyncChan chan datasync.ResyncEvent,
	keyPrefixes ...string) (datasync.WatchDataRegistration, error) {
	registrations := []datasync.WatchDataRegistration{}
	for _, transport := range ta.Adapters {
		watcherReg, err := transport.WatchData(resyncName, changeChan, resyncChan, keyPrefixes...)
		if err != nil {
			return nil, err
		}
		registrations = append(registrations, watcherReg)
	}

	return &AggregatedRegistration{
		Registrations: registrations,
	}, nil
}

// PublishData to every available transport
func (ta *TransportAggregator) PublishData(key string, data proto.Message) error {
	if len(ta.Adapters) == 0 {
		return fmt.Errorf("No transport is availabel in aggregator")
	}
	var wasError error
	for _, transport := range ta.Adapters {
		err := transport.PublishData(key, data)
		if err != nil {
			wasError = err
		}
	}
	return wasError
}

// Close every registration under watch aggregator
func (wa *AggregatedRegistration) Close() error {
	_, err := safeclose.CloseAll(wa.Registrations)
	return err
}
