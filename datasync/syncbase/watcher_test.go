// Copyright (c) 2018 Cisco and/or its affiliates.
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

package syncbase

import (
	"context"
	"github.com/ligato/cn-infra/datasync"
	"github.com/onsi/gomega"
	"testing"
)

// TestDeleteNonExisting verifies that delete operation for key with no
// prev value does not trigger change event
func TestDeleteNonExisting(t *testing.T) {

	const subPrefix = "/sub/prefix/"

	gomega.RegisterTestingT(t)

	var changes []datasync.ChangeEvent

	ctx, cancelFnc := context.WithCancel(context.Background())
	changeCh := make(chan datasync.ChangeEvent)
	resynCh := make(chan datasync.ResyncEvent)
	reg := NewRegistry()

	// register watcher
	wr, err := reg.Watch("resyncname", changeCh, resynCh, subPrefix)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(wr).NotTo(gomega.BeNil())

	// collect the change events
	go func() {
		for {
			select {
			case c := <-changeCh:
				changes = append(changes, c)
				c.Done(nil)
			case <-ctx.Done():
				break
			}
		}
	}()

	// execute the first set of changes
	changesToBePropagated := make(map[string]datasync.ChangeValue)

	// since the prev value does not exist this item should no trigger a change notification
	changesToBePropagated[subPrefix+"nonExistingDelete"] = NewChange(subPrefix+"nonExisting", nil, 0, datasync.Delete)
	// put should be propagated
	changesToBePropagated[subPrefix+"new"] = NewChange(subPrefix+"new", nil, 0, datasync.Put)

	err = reg.PropagateChanges(changesToBePropagated)
	gomega.Expect(err).To(gomega.BeNil())

	gomega.Expect(len(changes)).To(gomega.BeEquivalentTo(1))
	gomega.Expect(changes[0].GetKey()).To(gomega.BeEquivalentTo(subPrefix + "new"))
	gomega.Expect(changes[0].GetChangeType()).To(gomega.BeEquivalentTo(datasync.Put))

	// clear the changes
	changes = nil

	// remove an item that exist
	deleteItemThatExists := make(map[string]datasync.ChangeValue)
	deleteItemThatExists[subPrefix+"new"] = NewChange(subPrefix+"new", nil, 0, datasync.Delete)

	err = reg.PropagateChanges(deleteItemThatExists)
	gomega.Expect(err).To(gomega.BeNil())

	gomega.Expect(len(changes)).To(gomega.BeEquivalentTo(1))
	gomega.Expect(changes[0].GetKey()).To(gomega.BeEquivalentTo(subPrefix + "new"))
	gomega.Expect(changes[0].GetChangeType()).To(gomega.BeEquivalentTo(datasync.Delete))

	cancelFnc()

}
