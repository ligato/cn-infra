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
	"github.com/sirupsen/logrus"

	"go.ligato.io/cn-infra/v2/utils/redact"
)

func redactArgs(args ...interface{}) []interface{} {
	for i, arg := range args {
		args[i] = redact.Value(arg)
	}
	return args
}

type RedactHook struct {
}

func (r *RedactHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (r *RedactHook) Fire(e *logrus.Entry) error {
	return nil
}
