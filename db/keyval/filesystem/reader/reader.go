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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// API of file system reader
type API interface {
	// PathExists returns true if provided file exists, false otherwise
	PathExists(path string) bool
	// IsDirectory returns true if path is directory, false if it is a file
	IsDirectory(path string) (bool, error)
	// ProcessFile returns File type object. Provided path must be a file
	ProcessFile(path string) (File, error)
	// ProcessFilesInDir returns list of File type objects. Provided path must be a directory
	ProcessFilesInDir(path string) ([]File, error)
	// IsValid validates interface
	IsValid(ev fsnotify.Event) (bool, error)
}

// Reader implements file system API
type reader struct{}

// NewReader returns new instance of file reader
func NewReader() *reader {
	return &reader{}
}

// File data represents data structure of files used for configuration
type File struct {
	Path string
	Data []FileEntry `json:"data"`
}

// File data entry is single record of key-value, where key is defined as string, and value is modelled as raw message
// (rest of the json file under the "value").
type FileEntry struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

// Compares file with key-value set - new, modified and deleted entries. Result is against the parameter.
func (f *File) CompareTo(dataSet map[string][]byte) (changed, removed []FileEntry) {
	for key, value := range dataSet {
		var found bool
		for _, fData := range f.Data {
			if fData.Key == key {
				found = true
				if bytes.Compare(fData.Value, value) != 0 {
					changed = append(changed, fData)
					break
				}
			}
		}
		if !found {
			removed = append(removed, FileEntry{key, value})
		}
	}
	for _, fData := range f.Data {
		var found bool
		for key := range dataSet {
			if fData.Key == key {
				found = true
				break
			}
		}
		if !found {
			changed = append(changed, fData)
		}
	}

	return
}

func (r *reader) PathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// Validates file before reading, whether it is not a temporary file
func (r *reader) IsValid(ev fsnotify.Event) (bool, error) {
	// Silently skip empty event
	if ev.Name == "" {
		return false, nil
	}
	// Silently skip backup files ("~" at the end)
	if strings.Compare(ev.Name[len(ev.Name)-1:], "~") == 0 {
		return false, nil
	}
	// Validate file
	path := strings.Split(ev.Name, "/")
	if len(path) == 0 {
		return false, fmt.Errorf("error validating file path: invalid input %s", ev.Name)
	}
	fileName := path[len(path)-1]
	parts := strings.Split(fileName, ".")
	if len(parts) == 0 {
		return false, fmt.Errorf("error validating file in path %s: invalid format %s", ev.Name, fileName)
	}
	// Silently skip numeric files without extension (temporary files with PID of the editor)
	if len(parts) == 1 && r.isNumeric(parts[0]) {
		return false, nil
	}
	// Check file extension, skip
	fileExtension := parts[len(parts)-1]
	match, err := regexp.MatchString("sw.", fileExtension)
	if err != nil {
		return false, fmt.Errorf("error validating file %s: regex error: %v", ev.Name, err)
	}
	if match {
		return false, nil
	}
	return true, nil
}

// Read file from filesystem and un-marshall to required proto structure
func (r *reader) ProcessFile(path string) (File, error) {
	dataSet := File{}
	fileData, err := ioutil.ReadFile(path)
	if err != nil {
		return dataSet, fmt.Errorf("failed to read file from path %s: %v", path, err)
	}

	err = json.Unmarshal(fileData, &dataSet)
	if err != nil {
		return dataSet, fmt.Errorf("failed to unmarshal file %s: %v", path, err)
	}

	dataSet.Path = path
	return dataSet, nil
}

// Read all file names from directory and process them the ordinary way
func (r *reader) ProcessFilesInDir(path string) ([]File, error) {
	fileInfoList, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read files in directory %s", fileInfoList)
	}
	var files []File
	for _, fileInfo := range fileInfoList {
		file, err := r.ProcessFile(path + fileInfo.Name())
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

// Check if provided path is file, or directory
func (r *reader) IsDirectory(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open file/directory %s: %v", path, err)
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return false, fmt.Errorf("failed to get file info %s, %v", path, err)
	}
	return fileInfo.IsDir(), nil
}

func (r *reader) isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
