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

package supervisor

import (
	"io"
	"os"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"

	"github.com/pkg/errors"
)

// SvLogger is a logger object compatible with the process manager
type SvLogger struct {

	// standard I/O logger
	logStdio logging.Logger

	// log to file (optional)
	logFile logging.Logger

	// file reference
	file *os.File
}

// NewSvLogger prepares new supervisor logger for given process.
func NewSvLogger(name, logfilePath string, log logging.Logger) (svLogger *SvLogger, err error) {
	var file *os.File
	var logFile, logStdio logging.Logger

	if logfilePath != "" {
		if file, err = os.OpenFile(logfilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			return nil, errors.Errorf("failed to open log file %s: %v", logfilePath, err)
		}
		logFile = logrus.NewLogger(name)
		logFile.SetOutput(io.Writer(file))
	}

	logStdio = log
	logStdio.SetLevel(logging.DebugLevel)

	return &SvLogger{
		logStdio: log,
		logFile:  logFile,
		file:     file,
	}, nil
}

// Write message to the standard output and to a file if exists
func (l SvLogger) Write(p []byte) (n int, err error) {
	l.logStdio.Debugf("%s", p)
	if l.file != nil {
		l.logFile.Printf("%s\n", p)
	}
	return len(p), nil
}

// Close the file if necessary
func (l SvLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
