// Copyright (c) 2017 Cisco and/or its affiliates.
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

// Package idxmap provides mapping structure that allows to:
// - write & lookup data by indexes
// - index data using multiple indexes
// - watch data changes in the map
// - create local cache data of key value store (such as ETCD).
//
//  Primary Index                Item                                Secondary indexes
// ===================================================================================
//
//    Eth1              +---------------------+                 { "IP" : ["192.168.2.1", "10.0.0.8"],
//                      |  Status: Enabled    |                   "Type" : ["ethernet"]
//                      |  IP: 192.168.2.1    |                 }
//                      |      10.0.0.8       |
//                      |  Type: ethernet     |
//                      |  Desc: something    |
//                      +---------------------+
//
//
//
// Function `Put` adds a value (item) into the mapping. In the
// function call the primary index(name) for the item is specified. The
// values of the primary index are unique, if the name already exists,
// then the item is overwritten. To retrieve an item identified by the
// primary index, use the `Lookup` function. An item can be removed from
// the mapping by calling the `Delete` function. The names that
// are currently registered can be retrieved by calling the `ListNames`
// function.
//
// The constructor allows to define a `createIndexes` function that extracts
// secondary indices from stored items. The function returns a map indexed
// by names of secondary indexes, and the values are the extracted values
// for the particular item. The values of secondary indexes are not necessarily
// unique. To retrieve items based on secondary indicess use the
// `LookupByMetadata` function. In contrast to the lookup by primary index,
// the function may return multiple names.
//
// `Watch` allows to define a callback that is called when a change in the
// mapping occurs. There is a helper function `ToChan` available, which allows
// to deliver notifications through a channel.
package idxmap
