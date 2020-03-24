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

package redact

import (
	"fmt"
	"reflect"
)

var fieldPathsMap = map[reflect.Type][][]string{}

func RegisterFieldsPaths(x interface{}, paths ...[]string) {
	t := reflect.TypeOf(x)
	fieldPathsMap[t] = paths
}

func Mapped(x interface{}) {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("Redact used with non-Ptr kind %v", v.Kind()))
	}

	v = v.Elem()
	m, ok := fieldPathsMap[v.Type()]
	if !ok {
		return
	}
	for _, p := range m {
		r := v
		for _, f := range p {
			if r = r.FieldByName(f); r.IsZero() {
				panic("fieldPathsMap not found")
			}
		}
		if r.Kind() != reflect.String {
			panic("only strings can be mapped")
		}
		r.SetString(String(r.String()))
	}
	return
}
