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

package reader

import (
	"bytes"

	"github.com/fsnotify/fsnotify"
)

// API of file system reader
type API interface {
	// IsValid returns true if provided path is a file with extension reader can process. Returns false for
	// other files and directories
	IsValid(path string) bool
	// PathExists returns true if provided file exists, false otherwise.
	PathExists(path string) bool
	// ProcessFiles reads a file, or all the files if provided path is an directory, and leverages all data from them.
	// If there are multiple file extensions within the directory, only known files will be processed.
	ProcessFiles(path string) ([]*File, error)
	// Watch creates new file watcher and registers all provided paths. Also methods 'onEvent' and 'onClose' have to
	// be specified, and they are called with proper event/reader parameter whenever new notification from file system
	// arrives.
	Watch(paths []string, onEvent func(event fsnotify.Event, reader API), onClose func()) error
	// ToString returns string representation (name) of the reader
	ToString() string
	// Close releases all the reader resources. It also closes the file system channel which invokes watcher's 'onClose'
	// method
	Close() error
}

// File represents structure of file from system - path and set od key-value entries
type File struct {
	Path string
	Data []*DataEntry
}

// DataEntry is data record structure of a key-value
type DataEntry struct {
	Key   string
	Value []byte
}

// CompareTo compares file with key-value set - new, modified and deleted entries. Result is against the parameter.
func (f1 *File) CompareTo(f2 *File) (changed, removed []*DataEntry) {
	for _, f2Data := range f2.Data {
		var found bool
		for _, f1Data := range f1.Data {
			if f1Data.Key == f2Data.Key {
				found = true
				if bytes.Compare(f1Data.Value, f2Data.Value) != 0 {
					changed = append(changed, f1Data)
					break
				}
			}
		}
		if !found {
			removed = append(removed, &DataEntry{f2Data.Key, f2Data.Value})
		}
	}
	for _, f1Data := range f1.Data {
		var found bool
		for _, f2Data := range f2.Data {
			if f1Data.Key == f2Data.Key {
				found = true
				break
			}
		}
		if !found {
			changed = append(changed, f1Data)
		}
	}

	return
}
