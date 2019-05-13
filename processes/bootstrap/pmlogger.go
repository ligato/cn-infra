// Copyright (c) 2019 Cisco and/or its affiliates.
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

// Package vpp-agent implements the main entry point into the VPP Agent
// and it is used to build the VPP Agent executable.

package bootstrap

import (
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"sync"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
)

const defaultLogfilePath = "/var/log/vpp-agent-init.log"

var mx sync.Mutex

// Process manager logger, a helper struct which can be passed to the process
// manager writer
type pmLogger struct {
	log         logging.Logger
	logfilePath string
}

// Create a new process manager logger with default logfile path (if no custom is specified)
func newPmLogger(name, logfilePath string) *pmLogger {
	if logfilePath == "" {
		logfilePath = defaultLogfilePath
	}

	pmLog := &pmLogger{
		log:         logrus.NewLogger(name),
		logfilePath: logfilePath,
	}
	pmLog.log.SetOutput(os.Stdout)
	pmLog.log.SetLevel(logging.DebugLevel)
	return pmLog
}

// Write to logfile and print given message
func (l pmLogger) Write(p []byte) (n int, err error) {
	mx.Lock()
	defer mx.Unlock()

	file, err := os.OpenFile(l.logfilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return 0, errors.Errorf("error opening file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			l.log.Errorf("error closing file %s", l.logfilePath)
		}
	}()

	writer := io.Writer(file)
	log.SetOutput(writer)
	log.Printf("%s\n", p)

	l.log.Debugf("%s", p)

	return len(p), nil
}
