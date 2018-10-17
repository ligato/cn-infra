//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package filedb

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/pkg/errors"

	"github.com/ligato/cn-infra/db/keyval/filedb/reader"

	"github.com/fsnotify/fsnotify"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/logging"
)

// Client arranges communication between file system and internal database, and is also responsible for upstream events.
type Client struct {
	sync.Mutex
	log logging.Logger

	// A list of filesystem paths. It may be a specific file, or the whole directory. In case a path is a directory,
	// all files within are processed. If there are another directories inside, they are skipped.
	paths []string

	// Internal database mirrors changes in file system. Since the configuration can be only read, it is up to client
	// to handle difference between configuration revisions. Every database entry consists from three values:
	//  - path (where the configuration is written)
	//  - data key
	//  - data value
	// Note: database holds only configuration intended for agent with appropriate prefix
	db FilesSystemDB

	// File system reader API, grants access to methods needed for checking/reading of system files. Client expects
	// the reader to process files with specific extension
	readers []reader.API

	// A set of watchers for every key prefix.
	watchers map[string]chan keyedData

	// Prefix to recognize configuration for specific agent instance
	agentPrefix string
}

// NewClient initializes file watcher, database and registers paths provided via plugin configuration file
func NewClient(paths []string, prefix string, rd []reader.API, log logging.Logger) (*Client, error) {
	// Init client object
	c := &Client{
		paths:       paths,
		watchers:    make(map[string]chan keyedData),
		db:          NewDbClient(),
		readers:     rd,
		agentPrefix: prefix,
		log:         log,
	}

	// There can be various types of files, so client tries all available readers to obtain data from them into
	// a common list of files in paths. Initial read data are stored in internal database.
	for _, path := range paths {
		var filesInPath []*reader.File
		for _, r := range c.readers {
			files, err := r.ProcessFiles(path)
			if err != nil {
				return nil, fmt.Errorf("failed to process file/directory %s with %s: %v", path, r.ToString(), err)
			}
			filesInPath = append(filesInPath, files...)
		}

		// Find all the configuration for given agent and key and store it in the database
		for _, file := range filesInPath {
			for _, fileData := range file.Data {
				if ok := c.checkAgentPrefix(fileData.Key); ok {
					c.db.Add(file.Path, &reader.DataEntry{Key: fileData.Key, Value: fileData.Value})
				}
			}
		}
	}

	return c, nil
}

// GetPaths returns client file paths
func (c *Client) GetPaths() []string {
	return c.paths
}

// GetPrefix returns current agent prefix
func (c *Client) GetPrefix() string {
	return c.agentPrefix
}

// GetDataForFile returns data gor given file
func (c *Client) GetDataForFile(path string) []*reader.DataEntry {
	return c.db.GetDataForFile(path)
}

// GetDataForKey returns data gor given file
func (c *Client) GetDataForKey(key string) (*reader.DataEntry, bool) {
	return c.db.GetDataForKey(key)
}

// Represents data with both identifiers, file path and key.
type keyedData struct {
	path string
	watchResp
}

// BrokerWatcher implements CoreBrokerWatcher and provides broker/watcher constructors with client
type BrokerWatcher struct {
	*Client
	prefix string
}

// NewBroker provides BytesBroker object with client and given prefix
func (c *Client) NewBroker(prefix string) keyval.BytesBroker {
	return &BrokerWatcher{
		Client: c,
		prefix: prefix,
	}
}

// NewWatcher provides BytesWatcher object with client and given prefix
func (c *Client) NewWatcher(prefix string) keyval.BytesWatcher {
	return &BrokerWatcher{
		Client: c,
		prefix: prefix,
	}
}

// Put is not supported, fileDB plugin does not allow to do changes to the configuration
func (c *Client) Put(key string, data []byte, opts ...datasync.PutOption) error {
	c.log.Warnf("adding configuration to fileDB is currently not allowed")
	return nil
}

// NewTxn is not supported, filesystem plugin does not allow to do changes to the configuration
func (c *Client) NewTxn() keyval.BytesTxn {
	c.log.Warnf("creating transaction chains in fileDB is currently not allowed")
	return nil
}

// GetValue returns a value for given key
func (c *Client) GetValue(key string) (data []byte, found bool, revision int64, err error) {
	var entry *reader.DataEntry
	entry, found = c.db.GetDataForKey(c.agentPrefix + key)
	data = entry.Value
	return
}

// ListValues returns a list of values for given prefix
func (c *Client) ListValues(prefix string) (keyval.BytesKeyValIterator, error) {
	keyValues := c.db.GetDataForPrefix(c.agentPrefix + prefix)
	data := make([]*reader.DataEntry, 0, len(keyValues))
	for _, entry := range keyValues {
		data = append(data, &reader.DataEntry{
			Key:   c.stripAgentPrefix(entry.Key),
			Value: entry.Value,
		})
	}
	return &bytesKeyValIterator{len: len(data), data: data}, nil
}

// ListKeys returns a set of keys for given prefix
func (c *Client) ListKeys(prefix string) (keyval.BytesKeyIterator, error) {
	entries := c.db.GetDataForPrefix(c.agentPrefix + prefix)
	var keysWithoutPrefix []string
	for _, entry := range entries {
		keysWithoutPrefix = append(keysWithoutPrefix, c.stripAgentPrefix(entry.Key))
	}
	return &bytesKeyIterator{len: len(keysWithoutPrefix), keys: keysWithoutPrefix, prefix: prefix}, nil
}

// Delete is not allowed for fileDB, configuration file is read-only
func (c *Client) Delete(key string, opts ...datasync.DelOption) (existed bool, err error) {
	c.log.Warnf("deleting configuration from fileDB is currently not allowed")
	return false, nil
}

// Watch starts single watcher for every key prefix. Every watcher listens on its own data channel.
func (c *Client) Watch(resp func(response keyval.BytesWatchResp), closeChan chan string, keys ...string) error {
	c.Lock()
	defer c.Unlock()

	for _, key := range keys {
		dc := make(chan keyedData)
		go c.watch(resp, dc, closeChan, key)
		c.watchers[key] = dc
	}

	return nil
}

// Close closes all readers
func (c *Client) Close() error {
	if err := safeclose.Close(c.readers); err != nil {
		return errors.Errorf("failed to close file readers: %v", err)
	}
	return nil
}

// Awaits changes from data channel, prepares responses and sends them to the response function. Finally removes
// agent prefix from the key
func (c *Client) watch(resp func(response keyval.BytesWatchResp), dataChan chan keyedData, closeChan chan string, key string) {
	for {
		select {
		case keyedData, ok := <-dataChan:
			if !ok {
				return
			}
			keyedData.Key = c.stripAgentPrefix(keyedData.Key)
			if keyedData.Op == datasync.Delete {
				resp(&keyedData.watchResp)
			} else if bytes.Compare(keyedData.PrevValue, keyedData.Value) != 0 {
				resp(&keyedData.watchResp)
			}
		case <-closeChan:
			// TODO it seems this channel does not work
			return
		}
	}
}

// Processes events from file system. Every event is validated (all temporary or system files are omitted). Data
// is then read from the file as key-value pairs, and send to all data change watchers.

// Event watcher starts file system watcher for every reader available.
func (c *Client) eventWatcher() {
	for _, r := range c.readers {
		r.Watch(c.paths, c.onEvent, c.onClose)
	}
}

// OnEvent is common method called when new event from file system arrives. Different files may require different
// reader, but the data processing is the same.
func (c *Client) onEvent(event fsnotify.Event, r reader.API) {
	// If file was removed, delete all configuration associated with it. Do the same action for
	// rename, following action will be create with the new name which re-applies the configuration
	// (if new name is in scope of the defined path)
	if (event.Op == fsnotify.Rename || event.Op == fsnotify.Remove) && !r.PathExists(event.Name) {
		entries := c.db.GetDataForFile(event.Name)
		for _, entry := range entries {
			// Value from DB does not need to be checked
			keyed := keyedData{
				path:      event.Name,
				watchResp: watchResp{Op: datasync.Delete, Key: entry.Key, Value: nil, PrevValue: entry.Value},
			}
			c.sendToChannel(keyed)
			c.db.DeleteFile(event.Name)
		}
		return
	}

	// Read data from file
	dataSet, err := r.ProcessFiles(event.Name)
	if err != nil {
		c.log.Error(err)
		return
	}

	for _, fileData := range dataSet {
		// Compare with database, calculate diff
		latestFile := c.db.GetDataForFile(event.Name)
		changed, removed := fileData.CompareTo(&reader.File{
			Path: event.Name,
			Data: latestFile,
		})
		// Update database and propagate data to channel
		for _, data := range removed {
			if ok := c.checkAgentPrefix(data.Key); ok {
				keyed := keyedData{
					path:      fileData.Path,
					watchResp: watchResp{Op: datasync.Delete, Key: data.Key, Value: nil, PrevValue: data.Value},
				}
				c.sendToChannel(keyed)
				c.db.Delete(event.Name, keyed.Key)
			}
		}
		for _, data := range changed {
			if ok := c.checkAgentPrefix(data.Key); ok {
				// Get last key-val configuration item if exists
				var prevVal []byte
				if prevValEntry, ok := c.db.GetDataForPathAndKey(event.Name, data.Key); ok {
					prevVal = prevValEntry.Value
				}
				keyed := keyedData{
					path:      fileData.Path,
					watchResp: watchResp{Op: datasync.Put, Key: data.Key, Value: data.Value, PrevValue: prevVal},
				}
				c.sendToChannel(keyed)
				c.db.Add(event.Name, &reader.DataEntry{
					Key:   keyed.Key,
					Value: keyed.Value,
				})
			}
		}
	}
}

// OnClose is called from reader when the file system data channel is closed.
func (c *Client) onClose() {
	for _, channel := range c.watchers {
		close(channel)
	}
}

// Send data to correct channel. During the process, agent prefix is removed
func (c *Client) sendToChannel(keyed keyedData) {
	c.Lock()
	defer c.Unlock()

	for prefix, channel := range c.watchers {
		prefixedKey := keyed.Key
		if strings.HasPrefix(c.stripAgentPrefix(prefixedKey), prefix) {
			channel <- keyed
			return
		}
	}
}

// Verifies an agent prefix identifier
func (c *Client) checkAgentPrefix(key string) bool {
	if strings.HasPrefix(key, c.agentPrefix) {
		return true
	}
	return false
}

// Removes an agent prefix identifier from the key
func (c *Client) stripAgentPrefix(key string) string {
	return strings.Replace(key, c.agentPrefix, "", 1)
}
