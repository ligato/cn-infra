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

package main

import (
	"fmt"
	"github.com/gocql/gocql"
	"net"
	"gitlab.cisco.com/ctao/vnf-agent/examples/gocql/sql"
	"github.com/willfaught/gockle"
	"os"
	"github.com/ligato/cn-infra/db/sql/cassandra"
)

var UserTable = &User{}

type User struct {
	First_name string
	Last_name  string
	//NetIP      net.IP //mapped to native cassandra type
	//WrapIP  string //Wrapper01 used for custom (un)marshalling
	WrapIP2 *Wrapper01
	Udt03   *Udt03
	Udt04   Udt04
}

func main() {
	err := example()
	if err != nil {
		fmt.Println("failed ", err)
		os.Exit(1)
	} else {
		fmt.Println("sucessfull")
	}
}

func example() (err error) {
	// connect to the cluster
	cluster := gocql.NewCluster("172.17.0.1")
	//cluster.Keyspace = "demo"
	session, _ := cluster.CreateSession()

	defer session.Close()

	if err := session.Query("CREATE KEYSPACE IF NOT EXISTS demo WITH replication = {'class': 'SimpleStrategy', 'replication_factor' : 1};").
		Exec(); err != nil {
		return err
	}

	session.KeyspaceMetadata("demo")

	if err := session.Query(`CREATE TYPE IF NOT EXISTS udt03 (
		tx text,
		tx2 text)`).Exec(); err != nil {
		return err
	}
	if err := session.Query(`CREATE TYPE IF NOT EXISTS udt04 (
		ahoj text,
		caf frozen<udt03>)`).Exec(); err != nil {
		return err
	}

	if err := session.Query(`CREATE TABLE IF NOT EXISTS user (
			userid text PRIMARY KEY,
				first_name text,
				last_name text,
				Udt03 frozen<Udt03>,
				Udt04 frozen<Udt04>,
				NetIP inet,
				WrapIP text,
				emails set<text>,
				topscores list<int>,
				todo map<timestamp, text>
		);`).

		Exec(); err != nil {
		return err
	}

	if err := session.Query("CREATE INDEX IF NOT EXISTS demo_users_last_name ON user (last_name);").
		Exec(); err != nil {
		return err
	}

	_ /*ip01 */ , ipPrefix01, err := net.ParseCIDR("192.168.1.2/24")
	if err != nil {
		return err
	}
	db := cassandra.NewBrokerUsingSession(gockle.NewSession(session))
	err = db.Put("userid='Fero Mrkva'", &User{"Fero", "Mrkva", /*ip01, */
		//"kkk",
		&Wrapper01{ipPrefix01}, &Udt03{Tx: "tx1", Tx2: "tx2" /*, Inet1: "201.202.203.204"*/ },

		Udt04{"kuk", &Udt03{Tx: "txxxxxxxxx1", Tx2: "txxxxxxxxx2" /*, Inet1: "201.202.203.204"*/ }},
	})
	if err != nil {
		return err
	}

	users := &[]User{}
	err = sql.SliceIt(users, db.ListValues(sql.SelectFrom(users)+sql.Where(sql.FieldEq(&UserTable.Last_name, UserTable, "Mrkva"))))
	fmt.Println("users ", err, " ", users)

	return nil
}

// implements gocql.Marshaller, gocql.Unmarshaller
type Wrapper01 struct {
	ip *net.IPNet
}

func (w *Wrapper01) MarshalCQL(info gocql.TypeInfo) ([]byte, error) {

	if w.ip == nil {
		return []byte{}, nil
	}

	return []byte(w.ip.String()), nil
}
func (w *Wrapper01) UnmarshalCQL(info gocql.TypeInfo, data []byte) error {

	if len(data) > 0 {
		_, ipPrefix, err := net.ParseCIDR(string(data))

		if err != nil {
			return err
		}
		w.ip = ipPrefix
	}

	return nil
}

func (w *Wrapper01) String() string {
	if w.ip != nil {
		return w.ip.String()
	}

	return ""
}

type Udt03 struct {
	Tx  string `cql:"tx"`
	Tx2 string `cql:"tx2"`
	//Inet1 string
}

func (u *Udt03) String() string {
	return "{" + u.Tx + ", " + u.Tx2 /*+ ", " + u.Inet1*/ + "}"
}

type Udt04 struct {
	Ahoj string `cql:"ahoj"`
	Caf  *Udt03 `cql:"caf"`
	//Inet1 string
}

func (u *Udt04) String() string {
	return "{" + u.Ahoj + ", " + u.Caf.String() /*+ ", " + u.Inet1*/ + "}"
}
