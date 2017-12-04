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

package etcdv3

import (
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
)

// BytesWatchPutResp is sent when new key-value pair has been inserted
// or the value has been updated.
type BytesWatchPutRespItem struct {
	key       string
	value     []byte
	prevValue []byte
	rev       int64
}

type BytesWatchPutResp struct {
	response []*BytesWatchPutRespItem
}

// Watch starts subscription for changes associated with the selected <keys>.
// KeyPrefix defined in constructor is prepended to all <keys> in the argument
// list. The prefix is removed from the keys returned in watch events.
// Watch events will be delivered to <resp> callback.
func (pdb *BytesBrokerWatcherEtcd) Watch(resp func([]keyval.BytesWatchResp), closeChan chan string, keys ...string) error {
	var err error
	for _, k := range keys {
		err = watchInternal(pdb.Logger, pdb.watcher, closeChan, k, resp)
		if err != nil {
			break
		}
	}
	return err
}

func NewBytesWatchPutResp() *BytesWatchPutResp{
	var response []*BytesWatchPutRespItem
	return &BytesWatchPutResp{response}
}

// NewBytesWatchPutResp creates an instance of BytesWatchPutResp.
func NewBytesWatchPutRespItem(key string, value []byte, prevValue []byte, revision int64) *BytesWatchPutRespItem {
	return &BytesWatchPutRespItem{key: key, value: value, prevValue: prevValue, rev: revision}
}

// GetChangeType returns "Put" for BytesWatchPutResp.
func (resp *BytesWatchPutRespItem) GetChangeType() datasync.PutDel {
	return datasync.Put
}

// GetKey returns the key that the value has been inserted under.
func (resp *BytesWatchPutRespItem) GetKey() string {
	return resp.key
}

// GetValue returns the value that has been inserted.
func (resp *BytesWatchPutRespItem) GetValue() []byte {
	return resp.value
}

// GetPrevValue returns the previous value that has been inserted.
func (resp *BytesWatchPutRespItem) GetPrevValue() []byte {
	return resp.prevValue
}

// GetRevision returns the revision associated with the 'put' operation.
func (resp *BytesWatchPutRespItem) GetRevision() int64 {
	return resp.rev
}

type NewBytesWatchDelResp struct {
	response []*BytesWatchPutRespItem
}

// BytesWatchDelResp is sent when a key-value pair has been removed.
type BytesWatchDelRespItem struct {
	key string
	rev int64
}

// NewBytesWatchDelResp creates an instance of BytesWatchDelResp.
func NewBytesWatchDelRespItem(key string, revision int64) *BytesWatchDelRespItem {
	return &BytesWatchDelRespItem{key: key, rev: revision}
}

// GetChangeType returns "Delete" for BytesWatchPutResp.
func (resp *BytesWatchDelRespItem) GetChangeType() datasync.PutDel {
	return datasync.Delete
}

// GetKey returns the key that a value has been deleted from.
func (resp *BytesWatchDelRespItem) GetKey() string {
	return resp.key
}

// GetValue returns nil for BytesWatchDelResp.
func (resp *BytesWatchDelRespItem) GetValue() []byte {
	return nil
}

// GetPrevValue returns nil for BytesWatchDelResp.
func (resp *BytesWatchDelRespItem) GetPrevValue() []byte {
	return nil
}

// GetRevision returns the revision associated with the 'delete' operation.
func (resp *BytesWatchDelRespItem) GetRevision() int64 {
	return resp.rev
}
