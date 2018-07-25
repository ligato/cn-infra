// Copyright (c) 2018 Cisco and/or its affiliates.
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

package cryptodata

import (
	"github.com/ligato/cn-infra/db/keyval/kvproto"
	"github.com/golang/protobuf/proto"
	"strings"
	"reflect"
	"fmt"
)

type ProtoWrapperWrapper struct {
	kvproto.ProtoWrapper
	DecryptFunc DecryptFunc
	CryptoMap map[string][]string
}

func (db *ProtoWrapperWrapper) GetValue(key string, reqObj proto.Message) (found bool, revision int64, err error) {
	found, revision, err = db.ProtoWrapper.GetValue(key, reqObj)

	values, ok := db.CryptoMap[reqObj.String()]
	if !ok {
		return
	}

	for _, v := range values {
		reflected, err := getValueFromStruct(v, reqObj)
		if err != nil {
			return
		}

		if reflected.Kind() != reflect.String {
			return
		}

		reflectedString := reflected.String()
		decryptedBytes, err := db.DecryptFunc([]byte(reflectedString))
		if err != nil {
			return
		}

		reflected.SetString(string(decryptedBytes))
	}

	return
}

func getValueFromStruct(keyWithDots string, object interface{}) (*reflect.Value, error) {
    keySlice := strings.Split(keyWithDots, ".")
    v := reflect.ValueOf(object)

    for _, key := range keySlice[1:] {
        if v.Kind() == reflect.Ptr {
            v = v.Elem()
        }

        if v.Kind() != reflect.Struct {
            return nil, fmt.Errorf("only accepts structs; got %T", v)
        }

        v = v.FieldByName(key)
    }

	return &v, nil
}
