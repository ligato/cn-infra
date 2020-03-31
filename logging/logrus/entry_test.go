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
	"bytes"
	"fmt"
	"log/syslog"
	"testing"

	"github.com/sirupsen/logrus"
	syslog_hook "github.com/sirupsen/logrus/hooks/syslog"

	. "github.com/onsi/gomega"
)

func TestEntryPanicln(t *testing.T) {
	RegisterTestingT(t)

	errBoom := fmt.Errorf("boom time")

	defer func() {
		p := recover()
		Expect(p).NotTo(BeNil())

		switch pVal := p.(type) {
		case *logrus.Entry:
			Expect(pVal.Message).To(BeEquivalentTo("kaboom"))
			Expect(pVal.Data["err"]).To(BeEquivalentTo(errBoom))
		default:
			t.Fatalf("want type *LogMsg, got %T: %#v", pVal, pVal)
		}
	}()

	logger := NewLogger("testLogger")
	var buffer bytes.Buffer
	logger.SetOutput(&buffer)
	entry := NewEntry(logger)
	entry.WithField("err", errBoom).Panicln("kaboom")
}

func TestEntryPanicf(t *testing.T) {
	errBoom := fmt.Errorf("boom again")

	defer func() {
		p := recover()
		Expect(p).NotTo(BeNil())

		switch pVal := p.(type) {
		case *logrus.Entry:
			Expect("kaboom true").To(BeEquivalentTo(pVal.Message))
			Expect(errBoom).To(BeEquivalentTo(pVal.Data["err"]))
		default:
			t.Fatalf("want type *LogMsg, got %T: %#v", pVal, pVal)
		}
	}()

	logger := NewLogger("testLogger")
	var buffer bytes.Buffer
	logger.SetOutput(&buffer)
	entry := NewEntry(logger)
	entry.WithField("err", errBoom).Panicf("kaboom %v", true)
}

func TestAddHook(t *testing.T) {
	RegisterTestingT(t)

	logRegistry := NewLogRegistry()
	lgA := logRegistry.NewLogger("logger")
	Expect(lgA).NotTo(BeNil())

	hook, _ := syslog_hook.NewSyslogHook("", "", syslog.LOG_INFO, "")

	lgA.AddHook(hook)
	lgA.Info("Test Hook")
}
