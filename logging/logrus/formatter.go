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

package logrus

import (
	"strings"

	"github.com/sirupsen/logrus"
)

var defaultFormatter = &Formatter{
	TextFormatter: &logrus.TextFormatter{
		EnvironmentOverrideColors: true,
		TimestampFormat:           "2006-01-02 15:04:05.00000",
		/*CallerPrettyfier: func(frame *runtime.Frame) (funcVal string, fileVal string) {
			pcs := make([]uintptr, 20)
			depth := runtime.Callers(2, pcs)
			frames := runtime.CallersFrames(pcs[:depth])

			fmt.Println()
			for f, again := frames.Next(); again; f, again = frames.Next() {
				//pkg := getPackageName(f.Function)
				fmt.Printf("\t -> CALLER: %v\n", f.Function)
			}

			funcVal = frame.Function
			dir, file := filepath.Split(frame.File)
			fileVal = fmt.Sprintf("%s/%s:%d", filepath.Base(dir), file, frame.Line)
			return
		},*/
	},
}

// DefaultFormatter returns a formatter used as the default formatter for loggers.
func DefaultFormatter() *Formatter {
	return defaultFormatter
}

type Formatter struct {
	*logrus.TextFormatter
}

func (t *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	return t.TextFormatter.Format(entry)
}

// getPackageName reduces a fully qualified function name to the package name
// There really ought to be to be a better way...
func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}
