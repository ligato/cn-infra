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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/fsnotify/fsnotify"
	"github.com/ligato/cn-infra/logging/logrus"
)

const (
	// Reader name
	jsonYamlReader = "json-yaml-reader"
	// Supported extensions, other are ignored
	jsonExt = ".json"
	yamlExt = ".yaml"
)

// Reader is basic implementation of reader API, currently supporting JSON and YAML file types
type Reader struct {
	// Filesystem notification watcher
	watcher *fsnotify.Watcher
}

// Represents data structure of json/yaml files used for configuration
type dataFile struct {
	Data []dataFileEntry `json:"data"`
}

// Single record of key-value, where key is defined as string, and value is modelled as raw message
// (rest of the json/yaml file under the "value").
type dataFileEntry struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

// IsValid verifies given path and returns true if JsonReader is able to process it
func (r *Reader) IsValid(path string) bool {
	return r.isJSON(path) || r.isYAML(path)
}

// PathExists returns true if provided path exists within the filesystem, false otherwise
func (r *Reader) PathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// ProcessFiles reads all files within path, and un-marshals it to the common data representation.
func (r *Reader) ProcessFiles(path string) ([]*File, error) {
	files, err := r.getFilesFromPath(path)
	if err != nil {
		return nil, err
	}
	return r.getDataFromFiles(path, files)
}

// Write data to file
func (r *Reader) Write(path string, entry *DataEntry) error {
	files, err := r.getFilesFromPath(path)
	if err != nil {
		return err
	}
	fileDataSet, err := r.getDataFromFiles(path, files)
	// Expect empty or single entry
	file := &File{}
	if len(fileDataSet) == 1 {
		file = fileDataSet[0]
	} else if len(fileDataSet) > 1 {
		return fmt.Errorf("failed to write status data, unexpected result of target file processing")
	}
	// Add/update data
	var updated bool
	if len(fileDataSet) > 0 {
		for _, data := range fileDataSet[0].Data {
			if data.Key == entry.Key {
				data.Value = entry.Value
				updated = true
			}
		}
	}
	if !updated {
		file.Data = append(file.Data, entry)
	}

	return r.writeFile(path, file)
}

// Watch starts new filesystem notification watcher. All events from of json/yaml type files are passed to 'onEvent' function.
// Function 'onClose' is called when event channel is closed.
func (r *Reader) Watch(paths []string, onEvent func(event fsnotify.Event, reader API), onClose func()) error {
	var err error
	r.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to init json/yaml file watcher: %v", err)
	}
	for _, path := range paths {
		r.watcher.Add(path)
	}

	go func() {
		for {
			select {
			case event, ok := <-r.watcher.Events:
				if !ok {
					onClose()
					return
				}
				// Run event with proper reader
				if r.isJSON(event.Name) || r.isYAML(event.Name) {
					onEvent(event, r)
				}
			case err := <-r.watcher.Errors:
				if err != nil {
					logrus.DefaultLogger().Errorf("filesystem notification error %v", err)
				}
			}
		}
	}()

	return nil
}

// ToString returns reader's name
func (r *Reader) ToString() string {
	return jsonYamlReader
}

// Close closes the filesystem event watcher
func (r *Reader) Close() error {
	if r.watcher != nil {
		return r.watcher.Close()
	}
	return nil
}

// Get all files from provided path (single value if path is already a path, or a list of files if it is a directory).
// TODO nested directories are skipped.
func (r *Reader) getFilesFromPath(path string) (files []string, err error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info from %s: %v", path, err)
	}
	if fileInfo.IsDir() {
		fileInfoList, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read files in directory %s", fileInfoList)
		}
		for _, innerFileInfo := range fileInfoList {
			// Skip inner directories
			if !innerFileInfo.IsDir() {
				files = append(files, path+innerFileInfo.Name())
			}
		}
	} else {
		files = append(files, path)
	}

	return files, nil
}

// Reads every provided file a un-marshals it to a common file representation.
// Supports JSON and YAML
func (r *Reader) getDataFromFiles(path string, files []string) ([]*File, error) {
	var jsonYamlData []*File
	for _, file := range files {
		dataSet := dataFile{}
		fileData, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file from path %s: %v", file, err)
		}
		if r.isJSON(file) {
			err = json.Unmarshal(fileData, &dataSet)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal json file %s: %v", file, err)
			}
		} else if r.isYAML(file) {
			err = yaml.Unmarshal(fileData, &dataSet)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal yaml file %s: %v", file, err)
			}
		} else {
			continue
		}
		// Prepare data as byte array
		var fileDataSet []*DataEntry
		for _, data := range dataSet.Data {
			fileDataSet = append(fileDataSet, &DataEntry{
				Key:   data.Key,
				Value: data.Value,
			})
		}
		jsonYamlData = append(jsonYamlData, &File{
			Path: path,
			Data: fileDataSet,
		})
	}

	return jsonYamlData, nil
}

// Transforms data to tagged struct, marshals to JSON/YAML and writes to file
func (r *Reader) writeFile(path string, content *File) (err error) {
	var dataEntries []dataFileEntry
	for _, data := range content.Data {
		dataEntries = append(dataEntries, dataFileEntry{
			Key:   data.Key,
			Value: data.Value,
		})
	}
	data := &dataFile{
		Data: dataEntries,
	}
	var fileData []byte
	if r.isJSON(path) {
		if fileData, err = json.Marshal(data); err != nil {
			return fmt.Errorf("failed to marshall JSON %s: %v", path, err)
		}
	} else if r.isYAML(path) {
		if fileData, err = yaml.Marshal(data); err != nil {
			return fmt.Errorf("failed to marshall YAML %s: %v", path, err)
		}
	} else {
		return fmt.Errorf("failed to write data to %s with reader %s", path, jsonYamlReader)
	}
	file, err := os.OpenFile(path, os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to open status file %s for writing: %v", path, err)
	}
	defer file.Close()

	if _, err := file.Write(fileData); err != nil {
		return fmt.Errorf("failed to write status file %s for writing: %v", path, err)
	}
	return nil
}

func (r *Reader) isJSON(path string) bool {
	if strings.HasSuffix(path, jsonExt) {
		return true
	}
	return false
}

func (r *Reader) isYAML(path string) bool {
	if strings.HasSuffix(path, yamlExt) {
		return true
	}
	return false
}
