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
	"github.com/golang/protobuf/proto"
	"strings"
	"reflect"
	"fmt"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/pkg/errors"
	"encoding/base64"
)

type ProtoBrokerWrapper struct {
	keyval.ProtoBroker
	DecryptFunc DecryptFunc
	CryptoMap map[reflect.Type][]string
}

func (db *ProtoBrokerWrapper) GetValue(key string, reqObj proto.Message) (bool, int64, error) {
	found, revision, err := db.ProtoBroker.GetValue(key, reqObj)
	if !found || err != nil {
		return found, revision, err
	}

	values, ok := db.CryptoMap[reflect.TypeOf(reqObj)]
	if !ok {
		return found, revision, err
	}

	for _, v := range values {
		reflected, err := getValueFromStruct(v, reqObj)
		if err != nil {
			return found, revision, err
		}

		if reflected.Kind() != reflect.String {
			return found, revision, errors.New("reflected value is not string")
		}

		reflectedString := reflected.String()
		decodedReflectedString, err := base64.URLEncoding.DecodeString(reflectedString)
		if err != nil {
			return found, revision, err
		}

		decryptedBytes, err := db.DecryptFunc(decodedReflectedString)
		if err != nil {
			return found, revision, err
		}

		reflected.SetString(string(decryptedBytes))
	}

	return found, revision, err
}

func getValueFromStruct(keyWithDots string, object interface{}) (*reflect.Value, error) {
    keySlice := strings.Split(keyWithDots, ".")
    v := reflect.ValueOf(object)

    for _, key := range keySlice {
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
