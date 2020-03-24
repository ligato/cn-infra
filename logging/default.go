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

package logging

var (
	// DefaultLogger is the default logger
	DefaultLogger Logger

	// DefaultRegistry is the default logging registry
	DefaultRegistry Registry
)

func Debug(args ...interface{}) { DefaultLogger.Debug(args...) }

func Debugf(format string, args ...interface{}) { DefaultLogger.Debugf(format, args...) }

func Info(args ...interface{}) { DefaultLogger.Info(args...) }

func Infof(format string, args ...interface{}) { DefaultLogger.Infof(format, args...) }

func Warn(args ...interface{}) { DefaultLogger.Warn(args...) }

func Warnf(format string, args ...interface{}) { DefaultLogger.Warnf(format, args...) }

func Error(args ...interface{}) { DefaultLogger.Error(args...) }

func Errorf(format string, args ...interface{}) { DefaultLogger.Errorf(format, args...) }

func Fatal(args ...interface{}) { DefaultLogger.Fatal(args...) }

func Fatalf(format string, args ...interface{}) { DefaultLogger.Fatalf(format, args...) }
