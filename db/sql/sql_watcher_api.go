package sql

import (
	"github.com/ligato/cn-infra/db"
)

// Watcher define API for monitoring changes in a datastore
type Watcher interface {
	// Watch starts to monitor changes in data store. Watch events will be delivered to the callback.
	Watch(callback func(WatchResp), statement ...string) error
}

// WatchResp represents a notification about change. It is sent through the watch resp channel.
type WatchResp interface {
	GetChangeType() db.PutDel
	// GetValue returns the value in the event
	GetValue(outBinding interface{}) error
}

// ToChan TODO
func ToChan(respChan chan WatchResp) func(event WatchResp) {
	return func(WatchResp) {
		/*select {
		case respChan <- resp:
		case <-time.After(defaultOpTimeout):
			log.Warn("Unable to deliver watch event before timeout.")
		}

		select {
		case wresp := <-recvChan:
			for _, ev := range wresp.Events {
				handleWatchEvent(respChan, ev)
			}
		case <-closeCh:
			log.WithField("key", key).Debug("Watch ended")
			return
		}*/
	}
}
