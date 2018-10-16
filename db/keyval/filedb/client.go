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

	// A list of filesystem paths. It may be specific file, or the whole directory. In case a path is a directory,
	// all files within are processed.
	paths []string

	// Internal database mirrors changes in file system. Since the configuration can be only read, it is up to client
	// to handle difference between configuration revisions. Every database entry consists from three values:
	//  - path (where the configuration is written)
	//  - data key
	//  - data value
	// Note: database holds only configuration intended for agent with defined prefix
	db FilesSystemDB

	// File system reader, grants access to methods needed for checking/reading of system files
	r reader.API

	// Watcher watches over events incoming from the registered files.
	watcher  *fsnotify.Watcher
	watchers map[string]chan keyedData

	// Prefix to recognize configuration for specific agent instance
	agentPrefix string
}

// NewClient initializes file watcher, database and registers paths provided via plugin configuration file
func NewClient(paths []string, prefix string, rd reader.API, log logging.Logger) (*Client, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to init file watcher: %v", err)
	}

	// Init client object
	c := &Client{
		paths:       paths,
		watcher:     watcher,
		watchers:    make(map[string]chan keyedData),
		db:          NewDbClient(),
		r:           rd,
		agentPrefix: prefix,
		log:         log,
	}

	// Do the initial read and store everything to internal database
	for _, path := range paths {
		// Register to watcher
		if err := watcher.Add(path); err != nil {
			return nil, err
		}

		isDir, err := c.r.IsDirectory(path)
		if err != nil {
			return nil, err
		}
		// Read file, or all files if path is a directory
		var filesInPath []reader.File
		if isDir {
			filesInPath, err = c.r.ProcessFilesInDir(path)
			if err != nil {
				return nil, err
			}
		} else {
			fileInPath, err := c.r.ProcessFile(path)
			if err != nil {
				return nil, err
			}
			filesInPath = []reader.File{fileInPath}
		}
		// Find all the configuration for given agent and key and store it in the database
		for _, file := range filesInPath {
			for _, fileData := range file.Data {
				if ok := c.checkAgentPrefix(fileData.Key); ok {
					c.db.Add(file.Path, fileData.Key, fileData.Value)
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

// GetDB returns fileDB database
func (c *Client) GetDB() FilesSystemDB {
	return c.db
}

// GetWatcher returns fileDB watcher
func (c *Client) GetWatcher() *fsnotify.Watcher {
	return c.watcher
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
	data, found = c.db.GetDataForKey(c.agentPrefix + key)
	return
}

// ListValues returns a list of values for given prefix
func (c *Client) ListValues(prefix string) (keyval.BytesKeyValIterator, error) {
	keyValues := c.db.GetValuesForPrefix(c.agentPrefix + prefix)
	data := make([]*reader.FileEntry, len(keyValues))
	for key, value := range keyValues {
		data = append(data, &reader.FileEntry{
			Key:   c.stripAgentPrefix(key),
			Value: value,
		})
	}
	return &bytesKeyValIterator{len: len(data), data: data}, nil
}

// ListKeys returns a set of keys for given prefix
func (c *Client) ListKeys(prefix string) (keyval.BytesKeyIterator, error) {
	keys := c.db.GetKeysForPrefix(c.agentPrefix + prefix)
	var keysWithoutPrefix []string
	for _, key := range keys {
		keysWithoutPrefix = append(keysWithoutPrefix, c.stripAgentPrefix(key))
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

// Close closes the event watcher
func (c *Client) Close() error {
	if c.watcher != nil {
		return c.watcher.Close()
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
func (c *Client) eventWatcher() {
	go func() {
		for {
			select {
			case event, ok := <-c.watcher.Events:
				if !ok {
					c.log.Debugf("fileDB watcher closed")
					for _, channel := range c.watchers {
						close(channel)
					}
					return
				}
				// Let's validate the event (skip temporary files, etc.)
				if isValid, err := c.r.IsValid(event); err != nil {
					c.log.Error(err)
				} else if !isValid {
					continue
				}
				// If file was removed, delete all configuration associated with it. Do the same action for
				// rename, following action will be create with the new name which re-applies the configuration
				// (if new name is in scope of the defined path)
				if (event.Op == fsnotify.Rename || event.Op == fsnotify.Remove) && !c.r.PathExists(event.Name) {
					keyValues := c.db.GetDataFromFile(event.Name)
					for key, data := range keyValues {
						// Value from DB does not need to be checked
						keyed := keyedData{
							path:      event.Name,
							watchResp: watchResp{Op: datasync.Delete, Key: key, Value: nil, PrevValue: data},
						}
						c.sendToChannel(keyed)
						c.db.DeleteFile(event.Name)
					}
					continue
				}

				// Read data from file
				dataSet, err := c.r.ProcessFile(event.Name)
				if err != nil {
					c.log.Error(err)
					continue
				}

				// Compare with database, calculate diff
				latestFile := c.db.GetDataFromFile(event.Name)
				changed, removed := dataSet.CompareTo(latestFile)
				// Update database and propagate data to channel
				for _, data := range removed {
					if ok := c.checkAgentPrefix(data.Key); ok {
						keyed := keyedData{
							path:      dataSet.Path,
							watchResp: watchResp{Op: datasync.Delete, Key: data.Key, Value: nil, PrevValue: data.Value},
						}
						c.sendToChannel(keyed)
						c.db.Delete(event.Name, keyed.Key)
					}
				}
				for _, data := range changed {
					if ok := c.checkAgentPrefix(data.Key); ok {
						// Get last key-val configuration item if exists
						prevVal, _ := c.db.GetDataForPathAndKey(event.Name, data.Key)
						keyed := keyedData{
							path:      dataSet.Path,
							watchResp: watchResp{Op: datasync.Put, Key: data.Key, Value: data.Value, PrevValue: prevVal},
						}
						c.sendToChannel(keyed)
						c.db.Add(event.Name, keyed.Key, keyed.Value)
					}
				}
			case err := <-c.watcher.Errors:
				if err != nil {
					c.log.Errorf("error watching fileDB events: %v", err)
				}
			}
		}
	}()
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

// Verifies an agent identifier
func (c *Client) checkAgentPrefix(key string) bool {
	if strings.HasPrefix(key, c.agentPrefix) {
		return true
	}
	return false
}

// Removes an agent prefix from the key
func (c *Client) stripAgentPrefix(key string) string {
	return strings.Replace(key, c.agentPrefix, "", 1)
}
