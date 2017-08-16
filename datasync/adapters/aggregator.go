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

// TransportAggregator is cumulative adapter which contains all available transport types
type TransportAggregator struct {
	grpcAdapter datasync.TransportAdapter
	etcdAdapter datasync.TransportAdapter
	redisAdapter datasync.TransportAdapter
}


// RegisterGrpcTransport is used to register GRPC transport
func (ta *TransportAggregator) RegisterGrpcTransport(adapter datasync.TransportAdapter) {
	if adapter != nil {
		ta.grpcAdapter = adapter
		log.Info("GRPC transport adapter registered")
	}
}

// RegisterEtcdTransport is used to register ETCD transport
func (ta *TransportAggregator) RegisterEtcdTransport(adapter datasync.TransportAdapter) {
	if adapter != nil {
		ta.etcdAdapter = adapter
		log.Info("ETCD transport adapter registered")
	}
}

// RegisterRedisTransport is used to register Redis transport
func (ta *TransportAggregator) RegisterRedisTransport(adapter datasync.TransportAdapter) {
	if adapter != nil {
		ta.redisAdapter = adapter
		log.Info("Redis transport adapter registered")
	}
}

// RetrieveGrpcTransport returns GRPC transport if registered
func (ta *TransportAggregator) RetrieveGrpcTransport() datasync.TransportAdapter {
	if ta.grpcAdapter == nil {
		return nil
	}
	return ta.grpcAdapter
}

// RetrieveEtcdTransport returns ETCD transport if registered
func (ta *TransportAggregator) RetrieveEtcdTransport() datasync.TransportAdapter {
	if ta.etcdAdapter == nil {
		return nil
	}
	return ta.etcdAdapter
}

// RetrieveRedisTransport returns Redis transport if registered
func (ta *TransportAggregator) RetrieveRedisTransport() datasync.TransportAdapter {
	if ta.redisAdapter == nil {
		return nil
	}
	return ta.redisAdapter
}

// RetrieveTransport returns first available transport. If no transport is available, return error
func (ta *TransportAggregator) RetrieveTransport() (datasync.TransportAdapter, error) {
	if ta.etcdAdapter != nil {
		return ta.etcdAdapter, nil
	} else if ta.grpcAdapter != nil {
		return ta.grpcAdapter, nil
	} else {
		return nil, fmt.Errorf("No transport is available")
	}
}