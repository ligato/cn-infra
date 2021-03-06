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

package keyval

import (
	"time"

	"go.ligato.io/cn-infra/v2/datasync"
	"go.ligato.io/cn-infra/v2/logging"
)

// ProtoWatcher defines API for monitoring changes in datastore.
// Changes are returned as protobuf/JSON-formatted data.
type ProtoWatcher interface {
	// Watch starts monitoring changes associated with the keys.
	// Watch events will be delivered to callback (not channel) <respChan>.
	// Channel <closeChan> can be used to close watching on respective key
	Watch(respChan func(datasync.ProtoWatchResp), closeChan chan string, key ...string) error
}

// ToChanProto creates a callback that can be passed to the Watch function
// in order to receive JSON/protobuf-formatted notifications through a channel.
// If the notification cannot be delivered until timeout, it is dropped.
func ToChanProto(respCh chan datasync.ProtoWatchResp, opts ...interface{}) func(dto datasync.ProtoWatchResp) {
	return func(dto datasync.ProtoWatchResp) {
		select {
		case respCh <- dto:
			// success
		case <-time.After(datasync.DefaultNotifTimeout):
			logging.Warn("Unable to deliver notification")
		}
	}
}
