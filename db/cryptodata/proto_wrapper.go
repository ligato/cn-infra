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
	"reflect"
	"github.com/ligato/cn-infra/db/keyval"
	"encoding/base64"
	"fmt"
)

type ProtoBrokerWrapper struct {
	keyval.ProtoBroker
	DecryptFunc DecryptFunc
	CryptoMap   map[reflect.Type][][]string
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

	for _, path := range values {
		if err := db.decryptStruct(path, reqObj); err != nil {
			return found, revision, err
		}
	}

	return found, revision, err
}

func (db *ProtoBrokerWrapper) decryptStruct(path []string, object interface{}) (error) {
	v, ok := object.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(object)
	}

	for pathIndex, key := range path {
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Kind() == reflect.Struct {
			v = v.FieldByName(key)
		}

		if v.Kind() == reflect.Slice {
			for i := 0; i < v.Len(); i++ {
				if err := db.decryptStruct(path[pathIndex:], v.Index(i)); err != nil {
					return err
				}
			}

			return nil
		}

		if v.Kind() == reflect.String {
			decoded, err := base64.URLEncoding.DecodeString(v.String())
			if err != nil {
				return err
			}

			decrypted, err := db.DecryptFunc(decoded)
			if err != nil {
				return err
			}

			v.SetString(string(decrypted))
			return nil
		}

		return fmt.Errorf("failed to process path on %v", v)
	}

	return nil
}
