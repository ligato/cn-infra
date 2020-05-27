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
	"reflect"
	"runtime"
	"strings"
)

const (
	cninfraModulePath     = "go.ligato.io/cn-infra/v2/"
	sirupsenLogrusPkgPath = "sirupsen/logrus."

	maximumCallerDepth = 15
	minimumCallerDepth = 6
)

var packagePath string

func init() {
	type x struct{}
	packagePath = reflect.TypeOf(x{}).PkgPath()
}

// getCaller retrieves the name of the first non-logrus calling function
func getCaller() *runtime.Frame {
	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])
	for f, more := frames.Next(); more; f, more = frames.Next() {
		if strings.LastIndex(f.Function, sirupsenLogrusPkgPath) != -1 {
			continue
		}
		if strings.LastIndex(f.Function, packagePath) == -1 {
			return &f
		}
	}
	return nil
}
