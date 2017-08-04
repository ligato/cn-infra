package cassandra_test

import (
	"testing"

	"github.com/ligato/cn-infra/db/sql"
	"github.com/ligato/cn-infra/db/sql/cassandra"
	"github.com/onsi/gomega"
)

// TestPut1_convenient is most convenient way of putting one entity to cassandra
func TestPut1_convenient(t *testing.T) {
	gomega.RegisterTestingT(t)

	session := mockSession()
	defer session.Close()
	db := cassandra.NewBrokerUsingSession(session)

	sqlStr, _, _ := cassandra.PutExpToString(sql.FieldEQ(&JamesBond.ID), JamesBond)
	gomega.Expect(sqlStr).Should(gomega.BeEquivalentTo(
		"UPDATE User SET id = ?, first_name = ?, last_name = ? WHERE id = ?"))

	mockExec(session, sqlStr, []interface{}{
		"James Bond", //set ID
		"James",      //set first_name
		"Bond",       //set last_name
		"James Bond", //where
	})
	err := db.Put(sql.FieldEQ(&JamesBond.ID), JamesBond)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}

// TestPut2_EQ is most convenient way of putting one entity to cassandra
func TestPut2_EQ(t *testing.T) {
	gomega.RegisterTestingT(t)

	session := mockSession()
	defer session.Close()
	db := cassandra.NewBrokerUsingSession(session)

	sqlStr, _, _ := cassandra.PutExpToString(sql.FieldEQ(&JamesBond.ID), JamesBond)
	gomega.Expect(sqlStr).Should(gomega.BeEquivalentTo(
		"UPDATE User SET id = ?, first_name = ?, last_name = ? WHERE id = ?"))

	mockExec(session, sqlStr, []interface{}{
		"James Bond", //set ID
		"James",      //set first_name
		"Bond",       //set last_name
		"James Bond", //where
	})
	err := db.Put(sql.Field(&JamesBond.ID, sql.EQ(JamesBond.ID)), JamesBond)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}

// TestPut3_customTableSchema checks that generated SQL statements
// contain customized table name & schema (see interfaces sql.TableName, sql.SchemaName)
func TestPut3_customTableSchema(t *testing.T) {
	gomega.RegisterTestingT(t)

	session := mockSession()
	defer session.Close()
	db := cassandra.NewBrokerUsingSession(session)

	entity := &CustomizedTablenameAndSchema{ID: "id", LastName: "Bond"}

	sqlStr, _, _ := cassandra.PutExpToString(sql.FieldEQ(&entity.ID), entity)
	gomega.Expect(sqlStr).Should(gomega.BeEquivalentTo(
		"UPDATE my_custom_schema.my_custom_name SET id = ?, last_name = ? WHERE id = ?"))

	mockExec(session, sqlStr, []interface{}{
		"James Bond", //set ID
		"James",      //set first_name
		"Bond",       //set last_name
		"James Bond", //where
	})
	err := db.Put(sql.FieldEQ(&entity.ID), entity)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}
