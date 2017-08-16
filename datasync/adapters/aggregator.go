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
	"github.com/ligato/cn-infra/datasync"
	log "github.com/ligato/cn-infra/logging/logrus"
	"fmt"
)

type TransportAggregator struct {
	grpcAdapter datasync.TransportAdapter
	etcdAdapter datasync.TransportAdapter
}

func (t *TransportAggregator) RegisterGrpcTransport(adapter datasync.TransportAdapter) {
	if adapter != nil {
		t.grpcAdapter = adapter
		log.Warn("GRPC transport adapter registered")
	}
}

func (t *TransportAggregator) RegisterEtcdTransport(adapter datasync.TransportAdapter) {
	if adapter != nil {
		t.etcdAdapter = adapter
		log.Warn("ETCD transport adapter registered")
	}
}

func (t *TransportAggregator) RetrieveGrpcTransport() (datasync.TransportAdapter, error) {
	if t.grpcAdapter == nil {
		return nil, fmt.Errorf("GRPC adapter is not registered")
	}
	return t.grpcAdapter, nil
}

func (t *TransportAggregator) RetrieveEtcdTransport() (datasync.TransportAdapter, error) {
	if t.etcdAdapter == nil {
		return nil, fmt.Errorf("ETCD adapter is not registered")
	}
	return t.etcdAdapter, nil
}


