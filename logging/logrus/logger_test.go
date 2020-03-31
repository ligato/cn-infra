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
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"testing"

	. "github.com/onsi/gomega"
	lg "github.com/sirupsen/logrus"

	"go.ligato.io/cn-infra/v2/logging"
)

func logAndAssertJSON(t *testing.T, logFn func(*Logger), assertFn func(fields map[string]interface{})) {
	RegisterTestingT(t)

	logger := NewLogger("testLogger")
	logger.SetFormatter(&lg.JSONFormatter{})
	var buffer bytes.Buffer
	logger.SetOutput(&buffer)

	logFn(logger)

	var fields map[string]interface{}
	err := json.NewDecoder(&buffer).Decode(&fields)
	Expect(err).To(BeNil())

	assertFn(fields)
}

func logAndAssertText(t *testing.T, logFn func(*Logger), assertFn func(fields map[string]string)) {
	RegisterTestingT(t)

	logger := NewLogger("testLogger")
	logger.SetFormatter(&lg.TextFormatter{
		DisableColors: true,
	})
	var buffer bytes.Buffer
	logger.SetOutput(&buffer)

	logFn(logger)

	fields := make(map[string]string)
	for _, kv := range strings.Split(buffer.String(), " ") {
		if !strings.Contains(kv, "=") {
			continue
		}
		kvArr := strings.Split(kv, "=")
		key := strings.TrimSpace(kvArr[0])
		val := kvArr[1]
		if kvArr[1][0] == '"' {
			var err error
			val, err = strconv.Unquote(val)
			Expect(err).To(BeNil())
		}
		fields[key] = val
	}

	assertFn(fields)
}

func TestPrint(t *testing.T) {
	logAndAssertJSON(t, func(log *Logger) {
		log.Print("test")
	}, func(fields map[string]interface{}) {
		Expect(fields).To(And(
			HaveKeyWithValue("msg", "test"),
			HaveKeyWithValue("level", "info"),
		))
	})
}

func TestInfo(t *testing.T) {
	logAndAssertJSON(t, func(log *Logger) {
		log.Info("test")
	}, func(fields map[string]interface{}) {
		Expect(fields["msg"]).To(BeEquivalentTo("test"))
		Expect(fields["level"]).To(BeEquivalentTo("info"))
	})
}

func TestWarn(t *testing.T) {
	logAndAssertJSON(t, func(log *Logger) {
		log.Warn("test")
	}, func(fields map[string]interface{}) {
		Expect(fields["msg"]).To(BeEquivalentTo("test"))
		Expect(fields["level"]).To(BeEquivalentTo("warning"))
	})
}

func TestInfolnShouldAddSpacesBetweenStrings(t *testing.T) {
	logAndAssertJSON(t, func(log *Logger) {
		log.Infoln("test", "test")
	}, func(fields map[string]interface{}) {
		Expect(fields["msg"]).To(BeEquivalentTo("test test"))
	})
}

func TestInfolnShouldAddSpacesBetweenStringAndNonstring(t *testing.T) {
	logAndAssertJSON(t, func(log *Logger) {
		log.Infoln("test", 10)
	}, func(fields map[string]interface{}) {
		Expect(fields["msg"]).To(BeEquivalentTo("test 10"))
	})
}

func TestInfolnShouldAddSpacesBetweenTwoNonStrings(t *testing.T) {
	logAndAssertJSON(t, func(log *Logger) {
		log.Infoln(10, 10)
	}, func(fields map[string]interface{}) {
		Expect(fields["msg"]).To(BeEquivalentTo("10 10"))
	})
}

func TestInfoShouldAddSpacesBetweenTwoNonStrings(t *testing.T) {
	logAndAssertJSON(t, func(log *Logger) {
		log.Infoln(10, 10)
	}, func(fields map[string]interface{}) {
		Expect(fields["msg"]).To(BeEquivalentTo("10 10"))
	})
}

func TestInfoShouldNotAddSpacesBetweenStringAndNonstring(t *testing.T) {
	logAndAssertJSON(t, func(log *Logger) {
		log.Info("test", 10)
	}, func(fields map[string]interface{}) {
		Expect(fields["msg"]).To(BeEquivalentTo("test10"))
	})
}

func TestInfoShouldNotAddSpacesBetweenStrings(t *testing.T) {
	logAndAssertJSON(t, func(log *Logger) {
		log.Info("test", "test")
	}, func(fields map[string]interface{}) {
		Expect(fields["msg"]).To(BeEquivalentTo("testtest"))
	})
}

func TestWithFieldsShouldAllowAssignments(t *testing.T) {
	var buffer bytes.Buffer
	var fields logging.Fields

	logger := NewLogger("testLogger")
	logger.SetOutput(&buffer)
	logger.SetFormatter(new(lg.JSONFormatter))
	entry := NewEntry(logger)

	entry2 := entry.WithFields(logging.Fields{
		"key1": "value1",
	})

	entry2.WithField("key2", "value2").Info("test")
	err := json.NewDecoder(&buffer).Decode(&fields)
	Expect(err).To(BeNil())

	Expect("value2").To(BeEquivalentTo(fields["key2"]))
	Expect("value1").To(BeEquivalentTo(fields["key1"]))

	buffer = bytes.Buffer{}
	fields = logging.Fields{}
	entry2.Info("test")
	err = json.NewDecoder(&buffer).Decode(&fields)
	Expect(err).To(BeNil())

	_, ok := fields["key2"]
	Expect(ok).To(BeFalse())
	Expect(fields["key1"]).To(BeEquivalentTo("value1"))
}

func TestUserSuppliedFieldDoesNotOverwriteDefaults(t *testing.T) {
	logAndAssertJSON(t, func(log *Logger) {
		log.WithField("msg", "hello").Info("test")
	}, func(fields map[string]interface{}) {
		Expect(fields["msg"]).To(BeEquivalentTo("test"))
	})
}

func TestUserSuppliedMsgFieldHasPrefix(t *testing.T) {
	logAndAssertJSON(t, func(log *Logger) {
		log.WithField("msg", "hello").Info("test")
	}, func(fields map[string]interface{}) {
		Expect(fields["msg"]).To(BeEquivalentTo("test"))
		Expect(fields["fields.msg"]).To(BeEquivalentTo("hello"))
	})
}

func TestUserSuppliedTimeFieldHasPrefix(t *testing.T) {
	logAndAssertJSON(t, func(log *Logger) {
		log.WithField("time", "hello").Info("test")
	}, func(fields map[string]interface{}) {
		Expect(fields["fields.time"]).To(BeEquivalentTo("hello"))
	})
}

func TestUserSuppliedLevelFieldHasPrefix(t *testing.T) {
	logAndAssertJSON(t, func(log *Logger) {
		log.WithField("level", 1).Info("test")
	}, func(fields map[string]interface{}) {
		Expect(fields["level"]).To(BeEquivalentTo("info"))
		Expect(fields["fields.level"]).To(BeEquivalentTo(1.0)) // JSON has floats only
	})
}

func TestDefaultFieldsAreNotPrefixed(t *testing.T) {
	logAndAssertText(t, func(log *Logger) {
		ll := log.WithField("herp", "derp")
		ll.Info("hello")
		ll.Info("bye")
	}, func(fields map[string]string) {
		for _, fieldName := range []string{"fields.level", "fields.time", "fields.msg"} {
			if _, ok := fields[fieldName]; ok {
				t.Fatalf("should not have prefixed %q: %v", fieldName, fields)
			}
		}
	})
}

func TestDoubleLoggingDoesntPrefixPreviousFields(t *testing.T) {
	RegisterTestingT(t)

	var buffer bytes.Buffer
	var fields logging.Fields

	logger := NewLogger("testLogger")
	logger.SetOutput(&buffer)
	logger.SetFormatter(new(lg.JSONFormatter))

	entry := logger.WithField("context", "eating raw fish")

	entry.Info("looks delicious")

	err := json.Unmarshal(buffer.Bytes(), &fields)
	Expect(err).To(BeNil(), "should have decoded first message")
	Expect(fields).To(HaveLen(5),
		"should only have 6 fields (msg/time/level/context/loc/logger)")
	Expect(fields["msg"]).To(BeEquivalentTo("looks delicious"))
	Expect(fields["context"]).To(BeEquivalentTo("eating raw fish"))

	buffer.Reset()

	entry.Warn("omg it is!")

	err = json.Unmarshal(buffer.Bytes(), &fields)
	Expect(err).To(BeNil(), "should have decoded second message")
	Expect(len(fields)).To(BeEquivalentTo(5), "should only have 6 fields (msg/time/level/context/loc/logger)")
	Expect(fields["msg"]).To(BeEquivalentTo("omg it is!"))
	Expect(fields["context"]).To(BeEquivalentTo("eating raw fish"))
	Expect(fields["fields.msg"]).To(BeNil(), "should not have prefixed previous `msg` entry")
}

func TestGetSetLevelRace(t *testing.T) {
	logger := NewLogger("testLogger")

	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				logger.SetLevel(logging.InfoLevel)
			} else {
				logger.GetLevel()
			}
		}(i)

	}
	wg.Wait()
}

func TestLoggingRace(t *testing.T) {
	logger := NewLogger("testLogger")

	var wg sync.WaitGroup
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			logger.Info("info")
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestLogInterface(t *testing.T) {
	logFn := func(l *Logger) {
		b := l.WithField("key", "value")
		b.Info("Test")
	}

	// test logger
	logger := NewLogger("testLogger")
	var buffer bytes.Buffer
	logger.SetOutput(&buffer)
	logFn(logger)

	// test Entry
	e := logger.WithField("another", "value")
	logFn(e.(*Entry).logger)
}
