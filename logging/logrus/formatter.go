//  Copyright (c) 2020 Cisco and/or its affiliates.
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

package logrus

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

var defaultFormatter = NewFormatter()

// DefaultFormatter returns a formatter used as the default formatter for loggers.
func DefaultFormatter() *Formatter {
	return defaultFormatter
}

// Tag names for structured fields of log entry
const (
	LoggerKey   = "logger"
	FunctionKey = "func"
	LocationKey = "loc"
)

func sortKeys(keys []string) {
	sort.SliceStable(keys, func(i, j int) bool {
		if keys[j] == LocationKey && keys[i] != FunctionKey ||
			keys[j] == FunctionKey && keys[i] != LocationKey {
			return true
		}
		if keys[i] == LocationKey && keys[j] != FunctionKey ||
			keys[i] == FunctionKey && keys[j] != LocationKey {
			return false
		}
		return strings.Compare(keys[i], keys[j]) == -1
	})
}

type Formatter struct {
	Function bool
	Location bool
	FullPath bool

	Formatter logrus.Formatter
}

func NewFormatter() *Formatter {
	return &Formatter{
		Formatter: &logrus.TextFormatter{
			EnvironmentOverrideColors: true,
			TimestampFormat:           "2006-01-02 15:04:05.00000",
			SortingFunc:               sortKeys,
		},
	}
}

func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	if f.Function || f.Location {
		if caller := getCaller(); caller != nil {
			data := logrus.Fields{}
			if f.Function {
				data[FunctionKey] = caller.Function
			}
			if f.Location {
				file := caller.File
				if !f.FullPath {
					dir, name := filepath.Split(file)
					file = filepath.Join(filepath.Base(dir), name)
				}
				data[LocationKey] = fmt.Sprintf("%s:%d", file, caller.Line)
			}
			for k, v := range entry.Data {
				data[k] = v
			}
			entry.Data = data
		}
	}
	return f.Formatter.Format(entry)
}
