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

package logs_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"go.ligato.io/cn-infra/v2/logging"
	. "go.ligato.io/cn-infra/v2/logging/logs"
	"go.ligato.io/cn-infra/v2/utils/redact"
)

type Data struct {
	Username string
	Password string
}

func (d Data) Redacted() interface{} {
	d.Password = redact.String(d.Password)
	return d
}

func TestRedactLoggerDefault(t *testing.T) {
	testLoggerRedact(t, DefaultLogger())
}

func TestRedactLoggerCustom(t *testing.T) {
	testLoggerRedact(t, NewLogger("testlogger"))
}

func testLoggerRedact(t *testing.T, logger *Logger) {
	const pwd = "password123"
	data := Data{
		Username: "bob",
		Password: pwd,
	}

	tests := []struct {
		name string
		do   func()
	}{
		{"Log", func() { logger.Log(logrus.InfoLevel, data) }},
		{"Logln", func() { logger.Logln(logrus.InfoLevel, data) }},
		{"Logf", func() { logger.Logf(logrus.InfoLevel, "data: %+v", data) }},
		{"Print", func() { logger.Print(data) }},
		{"Println", func() { logger.Println(data) }},
		{"Printf", func() { logger.Printf("data: %+v", data) }},
		{"Trace", func() { logger.Trace(data) }},
		{"Traceln", func() { logger.Traceln(data) }},
		{"Tracef", func() { logger.Tracef("data: %+v", data) }},
		{"Debug", func() { logger.Debug(data) }},
		{"Debugln", func() { logger.Debugln(data) }},
		{"Debugf", func() { logger.Debugf("data: %+v", data) }},
		{"Info", func() { logger.Info(data) }},
		{"Infoln", func() { logger.Infoln(data) }},
		{"Infof", func() { logger.Infof("data: %+v", data) }},
		{"Warn", func() { logger.Warn(data) }},
		{"Warnln", func() { logger.Warnln(data) }},
		{"Warnf", func() { logger.Warnf("data: %+v", data) }},
		{"Warning", func() { logger.Warning(data) }},
		{"Warningln", func() { logger.Warningln(data) }},
		{"Warningf", func() { logger.Warningf("data: %+v", data) }},
		{"Error", func() { logger.Error(data) }},
		{"Errorln", func() { logger.Errorln(data) }},
		{"Errorf", func() { logger.Errorf("data: %+v", data) }},
		{"WithField", func() { logger.WithField("x", "y").Print(data) }},
		{"WithFields", func() { logger.WithFields(logging.Fields{"x": "y"}).Print(data) }},
		{"WithError", func() { logger.WithError(errors.New("error")).Print(data) }},
		{"WithContext", func() { logger.WithContext(context.TODO()).Print(data) }},
		{"WithTime", func() { logger.WithTime(time.Now()).Print(data) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger.SetLevel(logrus.TraceLevel)
			var b bytes.Buffer
			logger.SetOutput(&b)
			test.do()
			out := b.String()
			t.Log(out)
			if strings.Contains(out, pwd) {
				t.Fatalf("expected log output to not contain password (%q)", pwd)
			}
		})
	}
}

func TestRedactField(t *testing.T) {
	const pwd = "password123"
	data := Data{
		Username: "bob",
		Password: pwd,
	}

	logger := NewLogger("test")
	var b bytes.Buffer
	logger.SetOutput(&b)
	logger.WithField("data", data).Infof("data info")
	out := b.String()
	t.Log(out)
	if strings.Contains(out, pwd) {
		t.Fatalf("expected log output to not contain password (%q)", pwd)
	}
}
