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

package logrus

import (
	lg "github.com/Sirupsen/logrus"
	"github.com/ligato/cn-infra/logging"
)

type Entry struct {
	logger *Logger
	ent    *lg.Entry
}

func NewEntry(logger *Logger) *Entry {
	return &Entry{
		logger: logger,
		ent:    lg.NewEntry(logger.std),
	}
}

func (entry *Entry) GetTag() string {
	return entry.logger.GetTag()
}

func (entry *Entry) SetTag(tag ...string) {
	entry.logger.SetTag(tag...)
}

func (entry *Entry) ClearTag() {
	entry.logger.ClearTag()
}

func (entry *Entry) GetLineInfo(depth int) string {
	return entry.logger.GetLineInfo(depth)
}

func (entry *Entry) GetEntry() *lg.Entry {
	return entry.ent
}

func (entry *Entry) header() *Entry {
	return entry.logger.header(1)
}

// Returns the string representation from the reader and ultimately the
// formatter.
func (entry *Entry) String() (string, error) {
	return entry.ent.String()
}

// Add an error as single field (using the key defined in ErrorKey) to the Entry.
func (entry *Entry) WithError(err error) *Entry {
	return entry.withField(ErrorKey, err, 1)
}

// Add a single field to the Entry.
func (entry *Entry) withField(key string, value interface{}, depth ...int) *Entry {
	d := 1
	if depth != nil && len(depth) > 0 {
		d += depth[0]
	}

	return entry.withFields(Fields{key: value}, d)
}

func (entry *Entry) WithField(key string, value interface{}) logging.LogWithLevel {
	return entry.withField(key, value)
}

// Add a map of fields to the Entry.
func (entry *Entry) withFields(fields Fields, depth ...int) *Entry {
	d := entry.logger.depth + 1
	if depth != nil && len(depth) > 0 {
		d += depth[0]
	}
	f := make(lg.Fields, len(fields))
	for k, v := range fields {
		f[k] = v
	}
	if _, ok := f[tagKey]; !ok {
		f[tagKey] = entry.GetTag()
	}
	if _, ok := f[locKey]; !ok {
		f[locKey] = entry.GetLineInfo(d)
	}
	e := entry.ent.WithFields(f)
	return &Entry{
		logger: entry.logger,
		ent:    e,
	}
}

func (entry *Entry) WithFields(fields map[string]interface{}) logging.LogWithLevel {
	return entry.withFields(Fields(fields))
}

func (entry *Entry) Debug(args ...interface{}) {
	entry.ent.Debug(args...)
}

func (entry *Entry) Print(args ...interface{}) {
	entry.ent.Print(args...)
}

func (entry *Entry) Info(args ...interface{}) {
	entry.ent.Info(args...)
}

func (entry *Entry) Warn(args ...interface{}) {
	entry.ent.Warn(args...)
}

func (entry *Entry) Warning(args ...interface{}) {
	entry.ent.Warning(args...)
}

func (entry *Entry) Error(args ...interface{}) {
	entry.ent.Error(args...)
}

func (entry *Entry) Fatal(args ...interface{}) {
	entry.ent.Fatal(args...)
}

func (entry *Entry) Panic(args ...interface{}) {
	entry.ent.Panic(args...)
}

// Entry Printf family functions

func (entry *Entry) Debugf(format string, args ...interface{}) {
	entry.ent.Debugf(format, args...)
}

func (entry *Entry) Infof(format string, args ...interface{}) {
	entry.ent.Infof(format, args...)
}

func (entry *Entry) Printf(format string, args ...interface{}) {
	entry.ent.Printf(format, args...)
}

func (entry *Entry) Warnf(format string, args ...interface{}) {
	entry.ent.Warnf(format, args...)
}

func (entry *Entry) Warningf(format string, args ...interface{}) {
	entry.ent.Warningf(format, args...)
}

func (entry *Entry) Errorf(format string, args ...interface{}) {
	entry.ent.Errorf(format, args...)
}

func (entry *Entry) Fatalf(format string, args ...interface{}) {
	entry.ent.Fatalf(format, args...)
}

func (entry *Entry) Panicf(format string, args ...interface{}) {
	entry.ent.Panicf(format, args...)
}

// Entry Println family functions

func (entry *Entry) Debugln(args ...interface{}) {
	entry.ent.Debugln(args...)
}

func (entry *Entry) Infoln(args ...interface{}) {
	entry.ent.Infoln(args...)
}

func (entry *Entry) Println(args ...interface{}) {
	entry.ent.Println(args...)
}

func (entry *Entry) Warnln(args ...interface{}) {
	entry.ent.Warnln(args...)
}

func (entry *Entry) Warningln(args ...interface{}) {
	entry.ent.Warningln(args...)
}

func (entry *Entry) Errorln(args ...interface{}) {
	entry.ent.Errorln(args...)
}

func (entry *Entry) Fatalln(args ...interface{}) {
	entry.ent.Fatalln(args...)
}

func (entry *Entry) Panicln(args ...interface{}) {
	entry.ent.Panicln(args...)
}
