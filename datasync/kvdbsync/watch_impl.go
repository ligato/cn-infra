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

package kvdbsync

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"go.ligato.io/cn-infra/v2/datasync"
	"go.ligato.io/cn-infra/v2/datasync/resync"
	"go.ligato.io/cn-infra/v2/datasync/syncbase"
	"go.ligato.io/cn-infra/v2/db/keyval"
	"go.ligato.io/cn-infra/v2/logging/logrus"
)

var (
	// ResyncAcceptTimeout defines timeout used for
	// sending resync event to registered watchers.
	ResyncAcceptTimeout = time.Second * 1
	// ResyncDoneTimeout defines timeout used during
	// resync after which resync will return an error.
	ResyncDoneTimeout = time.Second * 5
)

// WatchBrokerKeys implements go routines on top of Change & Resync channels.
type watchBrokerKeys struct {
	resyncReg  resync.Registration
	changeChan chan datasync.ChangeEvent
	resyncChan chan datasync.ResyncEvent
	prefixes   []string
	adapter    *watcher
}

type watcher struct {
	db   keyval.ProtoBroker
	dbW  keyval.ProtoWatcher
	base *syncbase.Registry
}

// WatchAndResyncBrokerKeys calls keyval watcher Watch() & resync Register().
// This creates go routines for each tuple changeChan + resyncChan.
func watchAndResyncBrokerKeys(resyncReg resync.Registration, changeChan chan datasync.ChangeEvent, resyncChan chan datasync.ResyncEvent,
	closeChan chan string, adapter *watcher, keyPrefixes ...string) (keys *watchBrokerKeys, err error) {
	keys = &watchBrokerKeys{
		resyncReg:  resyncReg,
		changeChan: changeChan,
		resyncChan: resyncChan,
		adapter:    adapter,
		prefixes:   keyPrefixes,
	}

	var wasErr error
	if err := keys.resyncRev(); err != nil {
		wasErr = err
	}
	if resyncReg != nil {
		go keys.watchResync(resyncReg)
	}
	if changeChan != nil {
		if err := keys.adapter.dbW.Watch(keys.watchChanges, closeChan, keys.prefixes...); err != nil {
			wasErr = err
		}
	}
	return keys, wasErr
}

func (keys *watchBrokerKeys) watchChanges(x datasync.ProtoWatchResp) {
	var prev datasync.LazyValue
	if datasync.Delete == x.GetChangeType() {
		_, prev = keys.adapter.base.LastRev().Del(x.GetKey())
	} else {
		_, prev = keys.adapter.base.LastRev().PutWithRevision(x.GetKey(),
			syncbase.NewKeyVal(x.GetKey(), x, x.GetRevision()))
	}

	ch := NewChangeWatchResp(context.Background(), x, prev)
	keys.changeChan <- ch
	// TODO NICE-to-HAVE publish the err using the transport asynchronously
}

// resyncReg.StatusChan == Started => resync
func (keys *watchBrokerKeys) watchResync(resyncReg resync.Registration) {
	for resyncStatus := range resyncReg.StatusChan() {
		if resyncStatus.ResyncStatus() == resync.Started {
			err := keys.resync()
			if err != nil {
				// We are not able to propagate it somewhere else.
				logrus.DefaultLogger().Errorf("getting resync data failed: %v", err)
				// TODO NICE-to-HAVE publish the err using the transport asynchronously
			}
		}
		resyncStatus.Ack()
	}
}

// ResyncRev fill the PrevRevision map. This step needs to be done even if resync is ommited
func (keys *watchBrokerKeys) resyncRev() error {
	for _, keyPrefix := range keys.prefixes {
		revIt, err := keys.adapter.db.ListValues(keyPrefix)
		if err != nil {
			return err
		}
		// if there are data for given prefix, register it
		for {
			data, stop := revIt.GetNext()
			if stop {
				break
			}
			logrus.DefaultLogger().Debugf("registering key found in KV: %q", data.GetKey())

			keys.adapter.base.LastRev().PutWithRevision(data.GetKey(),
				syncbase.NewKeyVal(data.GetKey(), data, data.GetRevision()))
		}
	}

	return nil
}

// Resync fills the resyncChan with the most recent snapshot (db.ListValues).
func (keys *watchBrokerKeys) resync() error {
	iterators := map[string]datasync.KeyValIterator{}
	for _, keyPrefix := range keys.prefixes {
		it, err := keys.adapter.db.ListValues(keyPrefix)
		if err != nil {
			return errors.WithMessagef(err, "list values for %s failed", keyPrefix)
		}
		iterators[keyPrefix] = NewIterator(it)
	}

	resyncEvent := syncbase.NewResyncEventDB(context.Background(), iterators)

	select {
	case keys.resyncChan <- resyncEvent:
		// ok
	case <-time.After(ResyncAcceptTimeout):
		logrus.DefaultLogger().Warn("Timeout of resync send!")
		return errors.New("resync not accepted in time")
	}

	select {
	case err := <-resyncEvent.DoneChan:
		if err != nil {
			return errors.WithMessagef(err, "resync returned error")
		}
	case <-time.After(ResyncDoneTimeout):
		logrus.DefaultLogger().Warn("Timeout of resync callback!")
	}

	return nil
}

// String returns resyncName.
func (keys *watchBrokerKeys) String() string {
	return keys.resyncReg.String()
}
