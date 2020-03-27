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
	"testing"

	. "github.com/onsi/gomega"
)

func TestListLoggers(t *testing.T) {
	RegisterTestingT(t)

	logRegistry := NewLogRegistry()
	loggers := logRegistry.ListLoggers()
	Expect(loggers).NotTo(BeNil())

	lg, found := loggers[globalName]
	Expect(found).To(BeTrue())
	Expect(lg).NotTo(BeNil())
}

func TestNewLogger(t *testing.T) {
	RegisterTestingT(t)

	const loggerName = "myLogger"

	logRegistry := NewLogRegistry()
	lg := logRegistry.NewLogger(loggerName)
	Expect(lg).NotTo(BeNil())

	loggers := logRegistry.ListLoggers()
	Expect(loggers).NotTo(BeNil())

	fromRegistry, found := loggers[loggerName]
	Expect(found).To(BeTrue())
	Expect(fromRegistry).NotTo(BeNil())
}

func TestGetSetLevel(t *testing.T) {
	RegisterTestingT(t)

	const level = "error"

	logRegistry := NewLogRegistry()

	// existing logger
	err := logRegistry.SetLevel(globalName, level)
	Expect(err).To(BeNil())

	loggers := logRegistry.ListLoggers()
	Expect(loggers).NotTo(BeNil())

	logger, found := loggers[globalName]
	Expect(found).To(BeTrue())
	Expect(logger).NotTo(BeNil())
	Expect(loggers[globalName]).To(BeEquivalentTo(level))

	currentLevel, err := logRegistry.GetLevel(globalName)
	Expect(err).To(BeNil())
	Expect(level).To(BeEquivalentTo(currentLevel))

	// non-existing logger
	err = logRegistry.SetLevel("unknown", level)
	Expect(err).To(BeNil()) // will be kept in logger level map in registry

	_, err = logRegistry.GetLevel("unknown")
	Expect(err).NotTo(BeNil())
}

func TestGetLoggerByName(t *testing.T) {
	RegisterTestingT(t)

	const (
		loggerA = "myLoggerA"
		loggerB = "myLoggerB"
	)

	logRegistry := NewLogRegistry()
	lgA := logRegistry.NewLogger(loggerA)
	Expect(lgA).NotTo(BeNil())

	lgB := logRegistry.NewLogger(loggerB)
	Expect(lgB).NotTo(BeNil())

	returnedA, found := logRegistry.Lookup(loggerA)
	Expect(found).To(BeTrue())
	Expect(returnedA).To(BeEquivalentTo(lgA))

	returnedB, found := logRegistry.Lookup(loggerB)
	Expect(found).To(BeTrue())
	Expect(returnedB).To(BeEquivalentTo(lgB))

	unknown, found := logRegistry.Lookup("unknown")
	Expect(found).To(BeFalse())
	Expect(unknown).To(BeNil())
}

func TestClearRegistry(t *testing.T) {
	RegisterTestingT(t)

	const (
		loggerA = "loggerA"
		loggerB = "loggerB"
	)

	logRegistry := NewLogRegistry()
	lgA := NewLogger(loggerA)
	Expect(lgA).NotTo(BeNil())

	lgB := NewLogger(loggerB)
	Expect(lgB).NotTo(BeNil())

	logRegistry.ClearRegistry()

	_, found := logRegistry.Lookup(loggerA)
	Expect(found).To(BeFalse())

	_, found = logRegistry.Lookup(loggerB)
	Expect(found).To(BeFalse())

	_, found = logRegistry.Lookup(globalName)
	Expect(found).To(BeTrue())
}
