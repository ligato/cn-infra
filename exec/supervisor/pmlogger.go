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

package supervisor

import (
	"io"
	"os"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/pkg/errors"
)

const (
	defaultLogfilePath = "/var/log/"

	logfileSuffix = "-file"
)

// pmLoggerManager holds information about process loggers
type pmLoggerManager struct {
	pathToFile map[string]*os.File
	pathToChan map[string]chan []byte
}

// Close all log channels which closes respective files
func (lm *pmLoggerManager) Close() error {
	for _, channel := range lm.pathToChan {
		close(channel)
	}
	return nil
}

// Create a new process manager logger with default logfile path (if no custom is specified)
func (lm *pmLoggerManager) newPmLogger(name, logfilePath string) (pml *pmLogger, err error) {
	if logfilePath == "" {
		logfilePath = defaultLogfilePath + name
	}

	// standard output logger
	logStdOut := logrus.NewLogger(name)
	logFile := logrus.NewLogger(name + logfileSuffix)

	var fileChan chan []byte
	file, ok := lm.pathToFile[logfilePath]
	if !ok {
		file, err = os.OpenFile(logfilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, errors.Errorf("error opening file: %v", err)
		}
		fileChan = make(chan []byte)
		lm.pathToFile[logfilePath] = file
		lm.pathToChan[logfilePath] = fileChan

		go func() {
			for {
				p, ok := <-fileChan
				if !ok {
					if err := file.Close(); err != nil {
						logStdOut.Errorf("error closing file %s", logfilePath)
					}
					return
				}

				logFile.Printf("%s\n", p)
				logStdOut.Debugf("%s", p)
			}
		}()

	} else {
		fileChan, ok = lm.pathToChan[logfilePath]
		if !ok {
			return nil, errors.Errorf("requested non-existing channel to %s: %v", logfilePath, err)
		}
	}

	// standard output logger
	logStdOut.SetOutput(os.Stdout)
	logStdOut.SetLevel(logging.DebugLevel)

	// file logger
	logFile.SetOutput(io.Writer(file))

	pmLog := &pmLogger{
		logStdOut: logStdOut,
		logFile:   logFile,
		fileChan:  fileChan,
	}
	return pmLog, nil
}

// Process manager logger, a helper struct which can be passed to the process
// manager writer
type pmLogger struct {
	logStdOut logging.Logger
	logFile   logging.Logger
	fileChan  chan []byte
}

func newPmLoggerManager() *pmLoggerManager {
	return &pmLoggerManager{
		pathToFile: make(map[string]*os.File),
		pathToChan: make(map[string]chan []byte),
	}
}

// Write to logfile and print given message
func (l pmLogger) Write(p []byte) (n int, err error) {
	l.fileChan <- p
	return len(p), nil
}
