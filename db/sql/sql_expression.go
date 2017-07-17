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

package sql

import (
	"bytes"
	"errors"
	"fmt"
	reflect2 "github.com/gocassa/gocassa/reflect"
	"reflect"
	"strings"
)

// Expression represents part of SQL statement and optional binding ("?")
type Expression interface {
	// Stringer prints default representation of SQL to String
	// Different implementations can override this using package specific func ExpToString()
	String() string

	// Binding are values referenced ("?") from the statement
	GetBinding() []interface{}

	// Accepts calls the methods on Visitor
	Accept(Visitor)
}

// Visitor for traversing expression tree
type Visitor interface {
	VisitPrefixedExp(*PrefixedExp)
	VisitFieldExpression(*FieldExpression)
}

// PrefixedExp TODO
type PrefixedExp struct {
	Prefix      string
	AfterPrefix Expression
	Suffix      string
	Binding     []interface{}
}

// String returns Prefix + " " + AfterPrefix
func (exp *PrefixedExp) String() string {
	if exp.AfterPrefix == nil {
		return exp.Prefix
	}
	return exp.Prefix + " " + exp.AfterPrefix.String()
}

// GetBinding is a getter...
func (exp *PrefixedExp) GetBinding() []interface{} {
	return exp.Binding
}

// Accept calls VisitPrefixedExp(...) & Accept(AfterPrefix)
func (exp *PrefixedExp) Accept(visitor Visitor) {
	visitor.VisitPrefixedExp(exp)
}

// FieldExpression TODO
type FieldExpression struct {
	PointerToAField interface{}
	AfterField      Expression
}

// String returns Prefix + " " + AfterPrefix
func (exp *FieldExpression) String() string {
	prefix := fmt.Sprint("<field on ", exp.PointerToAField, ">")
	if exp.AfterField == nil {
		return prefix
	}
	return prefix + " " + exp.AfterField.String()
}

// GetBinding is a getter...
func (exp *FieldExpression) GetBinding() []interface{} {
	return nil
}

// Accept calls VisitFieldExpression(...) & Accept(AfterField)
func (exp *FieldExpression) Accept(visitor Visitor) {
	visitor.VisitFieldExpression(exp)
}

// EvaluateUpdate TODO
func EvaluateUpdate(cfName string, val interface{} /*, opts Options*/) (
	statement string, fields []string, err error) {

	fields, _, ok := reflect2.FieldsAndValues(val)
	if !ok {
		return "", []string{}, errors.New("Not ok input val")
	}

	statement = updateStatement(cfName, fields)
	return statement, fields, nil
}

// SELECT TODO
func SELECT(entity interface{}, afterKeyword Expression, binding ...interface{}) Expression {
	return &PrefixedExp{"SELECT", FROM(entity, afterKeyword), "", binding}
}

// FROM TODO
func FROM(entity interface{}, afterKeyword Expression) Expression {
	return &PrefixedExp{"FROM", afterKeyword, "", []interface{}{entity}}
}

// SelectFields generates comma separated field names string
func SelectFields(val interface{} /*, opts Options*/) (statement string) {
	fields, _, ok := reflect2.FieldsAndValues(val)
	if !ok {
		return ""
	}

	return strings.Join(fields, ", ")
}

// WHERE keyword of SQL statement
func WHERE(afterKeyword Expression) Expression {
	return &PrefixedExp{"WHERE", afterKeyword, "", nil}
}

// DELETE keyword of SQL statement
func DELETE(entity interface{}, afterKeyword Expression) Expression {
	return &PrefixedExp{"DELETE", afterKeyword, "", nil}
}

/*
// Update generate SQL Condition
func Where(kn, cfName string, val interface{}) string {

}*/

// UPDATE keyspace.Movies SET col1 = val1, col2 = val2
func updateStatement(cfName string, fields []string /*, opts Options*/) (statement string) {
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf("UPDATE %s ", cfName))

	/*
		// Apply options
		if opts.TTL != 0 {
			buf.WriteString("USING TTL ")
			buf.WriteString(strconv.FormatFloat(opts.TTL.Seconds(), 'f', 0, 64))
			buf.WriteRune(' ')
		}*/

	buf.WriteString("SET ")
	first := true
	for _, fieldName := range fields {
		if !first {
			buf.WriteString(", ")
		} else {
			first = false
		}
		buf.WriteString(fieldName)
		buf.WriteString(` = ?`)
	}

	return buf.String()
}

// Exp rarely used parts of SQL statements
func Exp(statement string, binding ...interface{}) Expression {
	return &PrefixedExp{statement, nil, "", binding}
}

//TODO AND, OR

// Field is a helper function
//
// Example usage:
//   Where(Field(&UsersTable.LastName, UsersTable, EQ('Bond'))
//   // generates for example "WHERE last_name='Bond'"
func Field(pointerToAField interface{}, rigthOperand Expression) (exp Expression) {
	return &FieldExpression{pointerToAField, rigthOperand}
}

// FindField compares the pointers (pointerToAField with all fields in pointerToAStruct)
func FindField(pointerToAField interface{}, pointerToAStruct interface{}) (field *reflect.StructField, found bool) {
	fieldVal := reflect.ValueOf(pointerToAField)

	if fieldVal.Kind() != reflect.Ptr {
		panic("pointerToAField must be a pointer")
	}

	strct := reflect.Indirect(reflect.ValueOf(pointerToAStruct))
	numField := strct.NumField()
	for i := 0; i < numField; i++ {
		sf := strct.Field(i)

		if sf.CanAddr() {
			if fieldVal.Pointer() == sf.Addr().Pointer() {
				field := strct.Type().Field(i)
				return &field, true
			}
		}
	}

	return nil, false
}

// FieldName TODO
func FieldName(field *reflect.StructField) (name string, exported bool) {
	cql := field.Tag.Get("cql")
	if len(cql) > 0 {
		if cql == "-" {
			return cql, false
		}
		return cql, true
	}
	return field.Name, true
}

// EQ TODO
func EQ(binding interface{}) (exp Expression) {
	return &PrefixedExp{"=", Exp("?", binding), "", nil}
}

// Parenthesis TODO
func Parenthesis(inside Expression) (exp Expression) {
	return &PrefixedExp{"(", inside, ")", nil}
}

// IN TODO
func IN(binding ...interface{}) (exp Expression) {
	return &PrefixedExp{"IN(", nil, ")", binding}
}
