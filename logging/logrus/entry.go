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
	"fmt"
	lg "github.com/Sirupsen/logrus"
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
	lgEntry *lg.Entry
}

// NewEntry creates net entry object which stores provided logger and logrus' entry
func NewEntry(logger *Logger) *Entry { //todo
	lgEntry := lg.NewEntry(logger.std)
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

// WithError calls logrus 'WithError' method
func (entry *Entry) WithError(err error) *lg.Entry {
	return entry.lgEntry.WithField("error", err)
}

// WithField calls logrus 'WithField' method
func (entry *Entry) WithField(key string, value interface{}) *lg.Entry {
	return entry.lgEntry.WithFields(lg.Fields{key: value})
}

// WithFields calls logrus 'WithFields' method
func (entry *Entry) WithFields(fields lg.Fields) *lg.Entry {
	return entry.lgEntry.WithFields(fields)
}

// Debug logs a message at level Debug on the standard logger.
func (entry *Entry) Debug(args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.DebugLevel {
		entry.lgEntry.Debug(args...)
	}
}

// Print logs a message at level Info on the standard logger.
func (entry *Entry) Print(args ...interface{}) {
	entry.lgEntry.Info(args...)
}

// Info logs a message at level Info on the standard logger.
func (entry *Entry) Info(args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.InfoLevel {
		entry.lgEntry.Info(args...)
	}
}

// Warn logs a message at level Warn on the standard logger.
func (entry *Entry) Warn(args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.WarnLevel {
		entry.lgEntry.Warn(args...)
	}
}

// Warning logs a message at level Warn on the standard logger.
func (entry *Entry) Warning(args ...interface{}) {
	entry.Warn(args...)
}

// Error logs a message at level Error on the standard logger.
func (entry *Entry) Error(args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.ErrorLevel {
		entry.lgEntry.Error(args...)
	}
}

// Fatal logs a message at level Fatal on the standard logger.
func (entry *Entry) Fatal(args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.FatalLevel {
		entry.lgEntry.Fatal(args...)
	}
}

// Panic logs a message at level Panic on the standard logger.
func (entry *Entry) Panic(args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.PanicLevel {
		entry.lgEntry.Panic(args...)
	}
}

// Debugf logs a message at level Debug on the standard logger.
func (entry *Entry) Debugf(format string, args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.DebugLevel {
		entry.lgEntry.Debugf(format, args...)
	}
}

// Infof logs a message at level Info on the standard logger.
func (entry *Entry) Infof(format string, args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.InfoLevel {
		entry.lgEntry.Infof(format, args...)
	}
}

// Printf logs a message at level Info on the standard logger.
func (entry *Entry) Printf(format string, args ...interface{}) {
	entry.lgEntry.Printf(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func (entry *Entry) Warnf(format string, args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.WarnLevel {
		entry.lgEntry.Warnf(format, args...)
	}
}

// Warningf logs a message at level Warn on the standard logger.
func (entry *Entry) Warningf(format string, args ...interface{}) {
	entry.lgEntry.Warningf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func (entry *Entry) Errorf(format string, args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.ErrorLevel {
		entry.lgEntry.Errorf(format, args...)
	}
}

// Fatalf logs a message at level Debug on the standard logger.
func (entry *Entry) Fatalf(format string, args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.FatalLevel {
		entry.lgEntry.Fatalf(format, args...)
	}
}

// Panicf logs a message at level Panic on the standard logger.
func (entry *Entry) Panicf(format string, args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.PanicLevel {
		entry.lgEntry.Panicf(format, args...)
	}
}

// Debugln logs a message at level Debug on the standard logger.
func (entry *Entry) Debugln(args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.DebugLevel {
		entry.lgEntry.Debugln(entry.sprintlnn(args...))
	}
}

// Infoln logs a message at level Info on the standard logger.
func (entry *Entry) Infoln(args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.InfoLevel {
		entry.lgEntry.Infoln(entry.sprintlnn(args...))
	}
}

// Println logs a message at level Info on the standard logger.
func (entry *Entry) Println(args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.InfoLevel {
		entry.lgEntry.Println(entry.sprintlnn(args...))
	}
}

// Warnln logs a message at level Warn on the standard logger.
func (entry *Entry) Warnln(args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.WarnLevel {
		entry.lgEntry.Warnln(entry.sprintlnn(args...))
	}
}

// Warningln logs a message at level Warn on the standard logger.
func (entry *Entry) Warningln(args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.WarnLevel {
		entry.lgEntry.Warningln(entry.sprintlnn(args...))
	}
}

// Errorln logs a message at level Error on the standard logger.
func (entry *Entry) Errorln(args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.ErrorLevel {
		entry.lgEntry.Errorln(entry.sprintlnn(args...))
	}
}

// Fatalln logs a message at level Fatal on the standard logger.
func (entry *Entry) Fatalln(args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.FatalLevel {
		entry.lgEntry.Fatalln(entry.sprintlnn(args...))
	}
}

// Panicln logs a message at level Panic on the standard logger.
func (entry *Entry) Panicln(args ...interface{}) {
	if entry.lgEntry.Logger.Level >= lg.PanicLevel {
		entry.lgEntry.Panicln(entry.sprintlnn(args...))
	}
}

// Remove spaces, which are added between operands, regardless of their type
func (entry *Entry) sprintlnn(args ...interface{}) string {
	msg := fmt.Sprintln(args...)
	return msg[:len(msg)-1]
}
