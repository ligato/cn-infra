package sql

import (
	"bytes"
	"errors"
	"fmt"
	reflect2 "github.com/gocassa/gocassa/reflect"
	"reflect"
	"strings"
)

// Update generate SQL Statement
func Update(kn, cfName string, val interface{} /*, opts Options*/) (
	statement string, fields []string, err error) {

	fields, _, ok := reflect2.FieldsAndValues(val)
	if !ok {
		return "", []string{}, errors.New("Not ok input val")
	}

	statement = updateStatement(kn, cfName, fields)
	return statement, fields, nil
}

// SelectFrom generates
func SelectFrom(val interface{} /*, opts Options*/) (statement string) {
	return "select " + SelectFields(val) + From(val)
}

// From TODO
func From(val interface{} /*, opts Options*/) (statement string) {
	return " from " + reflect.TypeOf(val).Name()
}

// SelectFields generates comma separated field names string
func SelectFields(val interface{} /*, opts Options*/) (statement string) {
	fields, _, ok := reflect2.FieldsAndValues(val)
	if !ok {
		return ""
	}

	return strings.Join(fields, ", ")
}

// Where TODO
func Where(ands ...string) (statement string) {
	var x []string
	x = ands
	return " where " + strings.Join(x, " AND ")
}

// Delete TODO
func Delete(ands ...string) (statement string) {
	return " where " + AND(ands...)
}

// AND TODO
func AND(ands ...string) (statement string) {
	var x []string
	x = ands
	return strings.Join(x, " AND ")
}

// OR TODO
func OR(ors ...string) (statement string) {
	var x []string
	x = ors
	return strings.Join(x, " OR ")
}

// FieldName TODO
func FieldName(field reflect.StructField) (name string, exported bool) {
	cql := field.Tag.Get("cql")
	if len(cql) > 0 {
		if cql == "-" {
			return cql, false
		}
		return cql, true
	}
	return field.Name, true
}

// FieldEq is a helper function
//
// Example usage:
//   Where(FieldEq(&UsersTable.LastName, UsersTable, 'Bond')
//   // generates for example "WHERE last_name='Bond'"
func FieldEq(field interface{}, containerStruct interface{}, arg string) (statement string) {
	fieldVal := reflect.ValueOf(field)

	if fieldVal.Kind() != reflect.Ptr {
		panic("field must be a pointer")
	}

	strct := reflect.Indirect(reflect.ValueOf(containerStruct))
	numField := strct.NumField()
	for i := 0; i < numField; i++ {
		sf := strct.Field(i)
		if sf.CanAddr() && fieldVal.Pointer() == sf.Addr().Pointer() {
			name, exported := FieldName(strct.Type().Field(i))
			if !exported {
				panic("not exported field " + strct.Field(i).String())
			}
			return Eq(name, arg)
		}
	}

	panic("not found filed in struct")
}

// Eq TODO
func Eq(field string, arg string) (statement string) {
	statement += field
	statement += " = "
	statement += "'" + arg + "'"

	return statement
}

/*
// Update generate SQL Statement
func Where(kn, cfName string, val interface{}) string {

}*/

// UPDATE keyspace.Movies SET col1 = val1, col2 = val2
func updateStatement(kn, cfName string, fields []string /*, opts Options*/) (statement string) {
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf("UPDATE %s.%s ", kn, cfName))

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
