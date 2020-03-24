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
	lg "github.com/sirupsen/logrus"
)

// Tag names for structured fields of log message
const (
	locKey    = "loc"
	tagKey    = "tag"
	loggerKey = "logger"
)

// Entry is the logging entry. It has logrus' entry struct which is a final or intermediate Logrus logging entry
type Entry struct {
	logger  *Logger
	fields  lg.Fields
	lgEntry *lg.Entry
}

// NewEntry creates net entry object which stores provided logger and logrus' entry
func NewEntry(logger *Logger) *Entry {
	lgEntry := lg.NewEntry(logger.Logger)
	return &Entry{
		logger:  logger,
		lgEntry: lgEntry,
	}
}

// String returns the string representation from the reader or the formatter.
func (entry *Entry) String() (string, error) {
	serialized, err := entry.lgEntry.Logger.Formatter.Format(entry.lgEntry)
	if err != nil {
		return "", err
	}
	str := string(serialized)
	return str, nil
}

// WithError calls calls WithField with error key
func (entry *Entry) WithError(err error) *Entry {
	return entry.WithField("error", err)
}

// WithField calls transforms key/value to field and passes to WithFields
func (entry *Entry) WithField(key string, value interface{}) *Entry {
	return entry.WithFields(lg.Fields{key: value})
}

// WithFields stores field entries. These entries are used later when log method (Info, Debug, etc) is called
func (entry *Entry) WithFields(fields lg.Fields) *Entry {
	data := make(lg.Fields, len(entry.fields)+len(fields))
	for k, v := range entry.fields {
		data[k] = v
	}
	for k, v := range fields {
		data[k] = v
	}
	return &Entry{logger: entry.logger, fields: data, lgEntry: entry.lgEntry}
}

func (entry *Entry) Log(lvl lg.Level, args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lvl {
		entry.lgEntry.WithFields(entry.fields).Log(lvl, redactArgs(args...)...)
	}
}

func (entry *Entry) Logf(lvl lg.Level, f string, args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lvl {
		entry.lgEntry.WithFields(entry.fields).Logf(lvl, f, redactArgs(args...)...)
	}
}

func (entry *Entry) Logln(lvl lg.Level, args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lvl {
		entry.lgEntry.WithFields(entry.fields).Logln(lvl, redactArgs(args...)...)
	}
}

// Debug logs a message at level Debug on the standard logger.
func (entry *Entry) Debug(args ...interface{}) {
	entry.Log(lg.DebugLevel, args...)
}

// Print logs a message at level Info on the standard logger.
func (entry *Entry) Print(args ...interface{}) {
	entry.Log(lg.InfoLevel, args...)
}

// Info logs a message at level Info on the standard logger.
func (entry *Entry) Info(args ...interface{}) {
	entry.Log(lg.InfoLevel, args...)
}

// Warn logs a message at level Warning on the standard logger.
func (entry *Entry) Warn(args ...interface{}) {
	entry.Log(lg.WarnLevel, args...)
}

// Warning logs a message at level Warning on the standard logger.
func (entry *Entry) Warning(args ...interface{}) {
	entry.Log(lg.WarnLevel, args...)
}

// Error logs a message at level Error on the standard logger.
func (entry *Entry) Error(args ...interface{}) {
	entry.Log(lg.ErrorLevel, args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func (entry *Entry) Fatal(args ...interface{}) {
	entry.Log(lg.FatalLevel, args...)
}

// Panic logs a message at level Panic on the standard logger.
func (entry *Entry) Panic(args ...interface{}) {
	entry.Log(lg.PanicLevel, args...)
}

// Debugf logs a message at level Debug on the standard logger.
func (entry *Entry) Debugf(format string, args ...interface{}) {
	entry.Logf(lg.DebugLevel, format, args...)
}

// Infof logs a message at level Info on the standard logger.
func (entry *Entry) Infof(format string, args ...interface{}) {
	entry.Logf(lg.InfoLevel, format, args...)
}

// Printf logs a message at level Info on the standard logger.
func (entry *Entry) Printf(format string, args ...interface{}) {
	entry.Logf(lg.InfoLevel, format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func (entry *Entry) Warnf(format string, args ...interface{}) {
	entry.Logf(lg.WarnLevel, format, args...)
}

// Warningf logs a message at level Warn on the standard logger.
func (entry *Entry) Warningf(format string, args ...interface{}) {
	entry.Logf(lg.WarnLevel, format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func (entry *Entry) Errorf(format string, args ...interface{}) {
	entry.Logf(lg.ErrorLevel, format, args...)
}

// Fatalf logs a message at level Debug on the standard logger.
func (entry *Entry) Fatalf(format string, args ...interface{}) {
	entry.Logf(lg.FatalLevel, format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func (entry *Entry) Panicf(format string, args ...interface{}) {
	entry.Logf(lg.PanicLevel, format, args...)
}

// Debugln logs a message at level Debug on the standard logger.
func (entry *Entry) Debugln(args ...interface{}) {
	entry.Logln(lg.DebugLevel, args...)
}

// Infoln logs a message at level Info on the standard logger.
func (entry *Entry) Infoln(args ...interface{}) {
	entry.Logln(lg.InfoLevel, args...)
}

// Println logs a message at level Info on the standard logger.
func (entry *Entry) Println(args ...interface{}) {
	entry.Logln(lg.InfoLevel, args...)
}

// Warningln logs a message at level Warn on the standard logger.
func (entry *Entry) Warningln(args ...interface{}) {
	entry.Logln(lg.WarnLevel, args...)
}

// Errorln logs a message at level Error on the standard logger.
func (entry *Entry) Errorln(args ...interface{}) {
	entry.Logln(lg.ErrorLevel, args...)
}

// Fatalln logs a message at level Fatal on the standard logger.
func (entry *Entry) Fatalln(args ...interface{}) {
	entry.Logln(lg.FatalLevel, args...)
}

// Panicln logs a message at level Panic on the standard logger.
func (entry *Entry) Panicln(args ...interface{}) {
	entry.Logln(lg.PanicLevel, args...)
}
