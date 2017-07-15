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

package cassandra

import (
	r "reflect"
	"gitlab.cisco.com/ctao/vnf-agent/examples/gocql/sql"
	"github.com/willfaught/gockle"
)

// NewBrokerUsingSession is a constructor. Use it like this:
//
// session := gockle.NewSession(gocql.NewCluster("172.17.0.1"))
// defer db.Close()
// db := NewBrokerUsingSession(session)
// db.ListValues(...)
func NewBrokerUsingSession(gocqlSession gockle.Session) *BrokerCassa {
	return &BrokerCassa{session: gocqlSession}
}

// BrokerCassa implements interface db.Broker. This implementation simplifies work with gocql in the way
// that it is not need to write "SQL" queries. But the "SQL" is not really hidden, one can use it if needed.
// The "SQL" queries are generated from the go structures (see more details in Put, Delete, Key, GetValue, ListValues).
type BrokerCassa struct {
	session gockle.Session
}

// KeyValIterator is an iterator returned by ListValues call
type ValIterator struct {
	Delegate gockle.Iterator
}

/*
// NewTxn creates a new Data Broker transaction. A transaction can
// hold multiple operations that are all committed to the data
// store together. After a transaction has been created, one or
// more operations (put or delete) can be added to the transaction
// before it is committed.
func (pdb *BrokerCassa) NewTxn() keyval.ProtoTxn {
	return &Txn{txn: pdb.broker.NewTxn(), serializer: pdb.serializer}
}
// Put writes the provided key-value item into the data store.
//
// Returns an error if the item could not be written, ok otherwise.
func (pdb *BrokerCassa) Put(statement string, value interface{}, opts ...keyval.PutOption) error {
	return errors.New("Not supported")
}
*/

// Put writes the entity into the data store.
// Returns an error if the item could not be written, ok otherwise.
//
// Example usage:
//
//    err = db.Put("ID='James Bond'", &User{"James Bond", "James", "Bond"})
func (pdb *BrokerCassa) Put(where string, value interface{} /*TODO TTL, opts ...keyval.PutOption*/) error {
	statement, _, err := sql.Update(r.ValueOf(value).Type().Name() /*TODO extract method / make customizable*/ ,
		value                                                      /*, TODO TTL*/)
	if err != nil {
		return err
	}

	return pdb.session.Exec(statement+" WHERE "+where, structFieldPtrs(value)...)
}

// Exec runs statement (AS-IS) using gocql
func (pdb *BrokerCassa) Exec(statement string, binding ...interface{}) error {
	return pdb.session.Exec(statement, binding...)
}

//TODO delete one object or multiple
// alternative 1: like in gocassa table
// alternative 2: allow to pass query
//
// Delete removes from datastore key-value items stored under key.
func (pdb *BrokerCassa) Delete(where string) (existed bool, err error) {
	/*	statement, _, err := sql.Delete(pdb.KeyNamespace, pdb.TableName, where)
		if err != nil {
			return err
		}

		fmt.Println("xxx statement ", statement)

		// TODO stop using reflect.FieldsAndValues
		fields, values, _ := reflect.FieldsAndValues(value) //note, that type is cached inside
		return pdb.session.Exec(statement+" WHERE "+where, ptrs(value, values, fields)...)
	*/
	panic("not yet implemented")
}

// GetValue retrieves one key-value item from the datastore. The item
// is identified by the provided key.
//
// If the item was found, its value is un-marshaled and placed in
// the `reqObj` message buffer and the function returns found=true.
// If the object was not found, the function returns found=false.
// Function returns revision=revision of the latest modification
// If an error was encountered, the function returns an error.
func (pdb *BrokerCassa) GetValue(query string, reqObj interface{}) (found bool, err error) {
	it := pdb.ListValues(query)
	stop := it.GetNext(reqObj)
	return !stop, it.Close()
}

// ListValues retrieves an iterator for elements stored under the provided key.
// ListValues runs query (AS-IS) using gocql.
// Use utilities to:
// - generate query string
// - fill slice by values from iterator (SliceIt).
//
// Example usage:
//
// db := cassa.NewBrokerUsingSession(session)
// query := sql.SelectFrom(UserTable) + sql.Where(sql.FieldEq(&UserTable.LastName, UserTable, "Bond"))
// users := &[]User{}
// err := sql.SliceIt(users, db.ListValues(query))
//
func (pdb *BrokerCassa) ListValues(statement string) *ValIterator {
	it := pdb.session.ScanIterator(statement)
	return &ValIterator{it}
}

// GetNext returns the following item from the result set. If data was returned, found is set to true.
func (it *ValIterator) GetNext(val interface{}) (stop bool) {
	ptrs := structFieldPtrs(val)

	ok := it.Delegate.Scan(ptrs...)
	return !ok //if not ok than stop
}

// structFieldPtrs iterates struct fields and return slice of pointers to field values
func structFieldPtrs(val interface{}) []interface{} {
	rVal := r.Indirect(r.ValueOf(val))
	ptrs := []interface{}{}
	for i := 0; i < rVal.NumField(); i++ {
		field := rVal.Field(i)

		switch field.Kind() {
		case r.Chan, r.Func /*TODO func*/ , r.Map, r.Ptr, r.Interface, r.Slice:
			if field.IsNil() {
				p := r.New(field.Type().Elem())
				field.Set(p)
				ptrs = append(ptrs, p.Interface())
			} else {
				ptrs = append(ptrs, field.Interface())
			}
			//case r.Ptr, r.Interface, r.Array, r.Map, r.SliceIt, r.UnsafePointer, r.Chan /*TODO slice, map...*/ :
		default:
			if field.CanAddr() {
				ptrs = append(ptrs, field.Addr().Interface())
			} else if field.IsValid() {
				ptrs = append(ptrs, field.Interface())
			} else {
				panic("ivalid field")
			}
		}
	}
	return ptrs
}

// Close the iterator. Not the error is important (may occure during marshalling/un-marshalling)
func (it *ValIterator) Close() error {
	return it.Delegate.Close()
}
