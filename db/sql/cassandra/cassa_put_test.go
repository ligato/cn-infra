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

	mockPut(session, "UPDATE User SET id = ?, first_name = ?, last_name = ? WHERE id = ?",
		[]interface{}{
			"James Bond",
			"James",
			"Bond",
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

	mockPut(session, "UPDATE User SET id = ?, first_name = ?, last_name = ? WHERE id = ?",
		[]interface{}{
			"James Bond",
			"James",
			"Bond",
		})

	err := db.Put(sql.Field(&JamesBond.ID, sql.EQ(JamesBond.ID)), JamesBond)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}
