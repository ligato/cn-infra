package cassandra_test

import (
	"testing"

	"github.com/ligato/cn-infra/db/sql"
	"github.com/ligato/cn-infra/db/sql/cassandra"
	"github.com/onsi/gomega"
)

// TestDel1_convenient is most convenient way of deletening from cassandra
func TestDel1_convenient(t *testing.T) {
	gomega.RegisterTestingT(t)

	session := mockSession()
	defer session.Close()
	db := cassandra.NewBrokerUsingSession(session)

	mockPut(session, "DELETE FROM User WHERE id = ?",
		[]interface{}{
			"James Bond",
			"James",
			"Bond",
		})

	err := db.Delete(sql.FROM(JamesBond, sql.WHERE(sql.Field(&(JamesBond.ID), sql.EQ(JamesBond.ID)))))
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}
