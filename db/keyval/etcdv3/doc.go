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

// Package etcd contains abstraction on top of the key-value data store. The package uses etcd version 3.
//
// The entity that provides access to the data store is called BytesBrokerEtcd.
//
//      +-------------------+       crud/watch         ______
//      |  BytesBrokerEtcd  |          ---->          | ETCD |
//      +-------------------+        []byte           +------+
//
// To create a BytesBrokerEtcd use the following function
//
//   import  "github.com/ligato/cn-infra/db/keyval/etcd"
//
//   db := etcd.NewBytesBrokerEtcd(config)
//
// config is a path to a file with the following format:
//
//  key-file: <filepath>
//  ca-file: <filepath>
//  cert-file: <filepath>
//  insecure-skip-tls-verify: <bool>
//  insecure-transport: <bool>
//  dial-timout: <nanoseconds>
//  endpoints:
//    - <address_1>:<port>
//    - <address_2>:<port>
//    - ..
//    - <address_n>:<port>
//
// Connection to etcd is established using the provided config behind the scenes.
//
// Alternatively, you may connect to etcd by your self and initialize data broker with a given client.
//
//    db := etcd.NewBytesBrokerUsingClient(client)
//
// Created BytesBrokerEtcd implements Broker and Watcher interfaces. The example of usage can be seen below.
//
// To insert single key-value pair into etcd run:
//		db.Put(key, data)
// To remove a value identified by key:
//      db.Delete(key)
//
// In addition to single key-value pair approach, the transaction API is provided. Transaction
// executes multiple operations in a more efficient way than one by one execution.
//    // create new transaction
//    txn := db.NewTxn()
//
//    // add put operation into the transaction
//    txn.Put(key, value)
//
//    // add delete operation into the transaction
//    txn.Delete(key, value)
//
//    // try to commit the transaction
//    err := txn.Commit()
//
// To retrieve a value identified by key:
//    data, found, rev, err := db.GetValue(key)
//    if err == nil && found {
//       ...
//    }
//
// To retrieve all values matching a key prefix:
//    itr, err := db.ListValues(key)
//    if err != nil {
//       for {
//          data, allReceived, rev, err := itr.GetNext()
//          if allReceived {
//              break
//          }
//          if err != nil {
//              return err
//          }
//          process data...
//       }
//    }
//
// To retrieve values in specified key range:
//    itr, err := db.ListValues(key)
//    if err != nil {
//       for {
//          data, rev, allReceived := itr.GetNext()
//          if allReceived {
//              break
//          }
//          process data...
//       }
//    }
//
// To list keys without fetching the values
//    itr, err := db.ListKeys(prefix)
//    if err != nil {
//       for {
//          key, rev, allReceived := itr.GetNext()
//          if allReceived {
//              break
//          }
//          process key...
//       }
//    }
//
// To start watching changes in etcd:
//     respChan := make(chan keyval.BytesWatchResp, 0)
//     err = dbw.Watch(respChan, key)
//     if err != nil {
//         os.Exit(1)
//     }
//     for {
//          select {
//              case resp := <-respChan:
//                 switch resp.GetChangeType() {
//                 case data.Put:
//					   key := resp.GetKey()
//                     value := resp.GetValue()
//                     rev := resp.GetRevision()
//                 case data.Delete:
//                     ...
//                 }
//          }
//     }
//
//
// BytesBrokerEtcd also allows to create BytesPluginBrokerEtcd. BytesPluginBrokerEtcd instances share BytesBrokerEtcd's connection
// to the etcd. Another benefit gained by using BytesPluginBrokerEtcd is the option to setup a prefix. The prefix
// will be automatically prepended to all keys in the put/delete requests made from the BytesPluginBrokerEtcd. In case of
// get-like calls (GetValue, ListValues, ...) the prefixed is trimmed from key and returned value contains only part following the
// prefix in the key field.
//
//      +-----------------------+
//      | BytesPluginBrokerEtcd |
//      +-----------------------+
//              |
//              |
//               ----------------->   +-------------------+       crud/watch         ______
//			                          |  BytesBrokerEtcd  |       ---->             | ETCD |
//               ----------------->   +-------------------+        ([]byte)         +------+
//              |
//              |
//      +-----------------------+
//      | BytesPluginBrokerEtcd |
//      +-----------------------+
//
// To create a BytesPluginBrokerEtcd run
//    pdb := db.NewPluginBroker(prefix)
//
// BytesPluginBrokerEtcd implements Broker and Watcher interfaces thus the usage is the same as shown above.
//
// The package also provides a proto decorator that aims to simplify manipulation of proto modelled data.
// Proto decorator accepts arguments of type proto.message and the data are marshaled
// into []byte behind the scenes.
//
//
//      +-----------------+-------------------+       crud/watch         ______
//      |  ProtoDecorator |  ProtoBrokerEtcd  |       ---->             | ETCD |
//      +-----------------+-------------------+        ([]byte)         +------+
//        (proto.Message)
//
// The api of proto decorator is very similar to the ProtoBrokerEtcd. The difference is that arguments of type []byte
// are replaced by arguments of type proto.Message and in some case one of the return values is transformed to output argument.
//
// Example of decorator initialization
//
//    // db is BytesBrokerEtcd initialized as shown at the top of the page
//    protoBroker := etcd.NewProtoBrokerEtcd(db)
//
// The only difference in the Put/Delete functions is type of the argument, apart from that usage is the same as described above.
//
// Example of retrieving single key-value pair using proto decorator:
//   // if the value exists it is unmarshalled into the msg
//   found, rev, err := protoBroker.GetValue(key, msg)
//
//
// To retrieve all values matching the key prefix use
//   resp, err := protoDb.ListValues(path)
//   if err != nil {
//      os.Exit(1)
//   }
//
//   for {
//      // phonebook.Contact is a proto modelled structure (implementing proto.Message interface)
//      contact := &phonebook.Contact{}
//      // the value is unmarshaled into the contact variable
//      kv, allReceived  := resp.GetNext()
//      if allReceived {
//         break
//      }
//	err = kv.GetValue(contact)
//      if err != nil {
//          os.Exit(1)
//      }
//      ... use contact
//  }
package etcdv3
