package cassandra_test

import (
	"github.com/willfaught/gockle"
	"github.com/maraino/go-mock"
	"errors"
	reflect2 "github.com/gocassa/gocassa/reflect"
	"github.com/gocql/gocql"
)

// test data
var JamesBond = User{"James Bond", "James", "Bond"}
var PeterBond = User{"Peter Bond", "Peter", "Bond"}

// instance that represents users table (used in queries to define columns)
var UserTable = &User{}

//var UsersTypeInfo = map[string /*FieldName*/ ]gocql.TypeInfo{
//runtimeutils.GetFunctionName(UserTable.GetLastName): gocql.NewNativeType(0x03, gocql.TypeVarchar, ""),
//"LastName":                                          gocql.NewNativeType(0x03, gocql.TypeVarchar, ""),
//}

// User is simple structure for testing purposes
type User struct {
	ID        string `cql:"id"`
	FirstName string `cql:"first_name"`
	LastName  string `cql:"last_name"`
}

// simple structure that holds values of one row for mock iterator
type row struct {
	values []interface{}
	fields []string
}

// mockQuery is a helper for testing. It setups mock iterator
func mockQuery(sessionMock *gockle.SessionMock, query string, rows ...*row) {
	sessionMock.When("ScanIterator", query, mock.Any).Return(&IteratorMock{rows: rows})
	sessionMock.When("Close").Return()

}

// mockPut is a helper for testing. It setups mock iterator with any parameters/arguments
func mockPut(sessionMock *gockle.SessionMock, entity interface{}) {
	row := cells(entity)
	len := len(row.values) + 1
	params := make([]interface{}, len)
	for i := 0; i < len; i++ {
		params[i] = mock.Any
	}

	sessionMock.When("Exec", params...).Return(nil)
	sessionMock.When("Close").Return()
}

// cells is a helper that harvests all exported fields values
func cells(entity interface{}) (cellsInRow *row) {
	fields, values, _ := reflect2.FieldsAndValues(entity)
	return &row{values, fields}
}

// IteratorMock is a mock Iterator. See github.com/maraino/go-mock.
type IteratorMock struct {
	index   int
	rows    []*row
	closed  bool
	lastErr error
}

// Close implements Iterator.
func (m IteratorMock) Close() error {
	if m.closed {
		return errors.New("already closed")
	}
	return m.lastErr
}

// Scan implements Iterator.
func (m *IteratorMock) Scan(results ...interface{}) bool {
	if len(m.rows) > m.index {
		for i := 0; i < len(results) && i < len(m.rows[m.index].values); i++ {
			// TODO !!! types of fields
			typeInfo := gocql.NewNativeType(0x03, gocql.TypeVarchar, "")
			bytes, err := gocql.Marshal(typeInfo, m.rows[m.index].values[i])
			if err != nil {
				m.lastErr = err
				return false
			}
			err = gocql.Unmarshal(typeInfo, bytes, results[i])
			if err != nil {
				m.lastErr = err
				return false
			}
		}

		m.index++
		return true
	}
	return false
}

// ScanMap implements Iterator.
func (m *IteratorMock) ScanMap(results map[string]interface{}) bool {
	if len(m.rows) > m.index {
		for i := 0; i < len(m.rows[m.index].values) && i < len(m.rows[m.index].fields); i++ {
			key := m.rows[m.index].fields[i]
			value := m.rows[m.index].values[i]
			results[key] = value
		}

		m.index++
		return true
	}
	return false
}

func mockSession() (sessionMock *gockle.SessionMock) {
	sessionMock = &gockle.SessionMock{}
	sessionMock.When("Close").Return(nil)
	return sessionMock
}
