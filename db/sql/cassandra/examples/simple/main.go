package main

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/ligato/cn-infra/db/sql"
	"github.com/ligato/cn-infra/db/sql/cassandra"
	"github.com/ligato/cn-infra/utils/config"
	"github.com/willfaught/gockle"
	"net"
	"os"
	"errors"
)

// UserTable global variable reused when building queries/statements
var UserTable = &User{}

// User is simple structure used in automated tests
type User struct {
	FirstName string `cql:"first_name"`
	LastName  string `cql:"last_name"`
	//NetIP      net.IP //mapped to native cassandra type
	WrapIP *Wrapper01 //used for custom (un)marshalling
	Udt03  *Udt03
	Udt04  Udt04
	UdtCol []Udt03
}

// SchemaName demo schema name
func (entity *User) SchemaName() string {
	return "demo"
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("failed - configuration ", err)
		os.Exit(1)
	}

	session, err := cassandra.CreateSessionFromClientConfigAndKeyspace(cfg, false)
	defer session.Close()
	if err != nil {
		fmt.Println("failed - session1 ", err)
		os.Exit(1)
	}

	err = exampleKeyspace(session)
	if err != nil {
		fmt.Println("failed - keyspace ", err)
		os.Exit(1)
	}

	sessionWithKeyspace, err := cassandra.CreateSessionFromClientConfigAndKeyspace(cfg, true)
	defer sessionWithKeyspace.Close()
	if err != nil {
		fmt.Println("failed - session2 ", err)
		os.Exit(1)
	}
	err = example(sessionWithKeyspace)
	if err != nil {
		fmt.Println("failed - example ", err)
		os.Exit(1)
	}
}

func loadConfig() (cassandra.ClientConfig, error) {
	var cfg cassandra.ClientConfig
	if len(os.Args) < 2 {
		return cfg, errors.New("Configuration filename argument not specified")
	}

	configFileName := os.Args[1]
	err := config.ParseConfigFromYamlFile(configFileName, &cfg)
	return cfg, err
}

func exampleKeyspace(session *gocql.Session) (err error) {
	if err := session.Query("CREATE KEYSPACE IF NOT EXISTS demo WITH replication = {'class': 'SimpleStrategy', 'replication_factor' : 1};").
		Exec(); err != nil {
		return err
	}

	return nil
}

func example(session *gocql.Session) (err error) {
	err = exampleDDL(session)
	if err != nil {
		return err
	}
	err = exampleDML(session)
	if err != nil {
		return err
	}

	return nil
}

func exampleDDL(session *gocql.Session) (err error) {
	if err := session.Query("CREATE KEYSPACE IF NOT EXISTS demo WITH replication = {'class': 'SimpleStrategy', 'replication_factor' : 1};").
		Exec(); err != nil {
		return err
	}
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
				UdtCol list<frozen<Udt03>>,
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

	return nil
}

func exampleDML(session *gocql.Session) (err error) {
	_ /*ip01 */, ipPrefix01, err := net.ParseCIDR("192.168.1.2/24")
	if err != nil {
		return err
	}
	db := cassandra.NewBrokerUsingSession(gockle.NewSession(session))
	written := &User{"Fero", "Mrkva", /*ip01, */
		&Wrapper01{ipPrefix01}, &Udt03{Tx: "tx1", Tx2: "tx2" /*, Inet1: "201.202.203.204"*/},

		Udt04{"kuk", &Udt03{Tx: "txxxxxxxxx1", Tx2: "txxxxxxxxx2" /*, Inet1: "201.202.203.204"*/}},
		[]Udt03{{Tx: "txt1Col", Tx2: "txt2Col"}},
	}
	err = db.Put(sql.Exp("userid='Fero Mrkva'"), written)
	if err == nil {
		fmt.Println("Successfully written: ", written)
	} else {
		return err
	}

	users := &[]User{}
	err = sql.SliceIt(users, db.ListValues(sql.FROM(UserTable,
		sql.WHERE(sql.Field(&UserTable.LastName, sql.EQ("Mrkva"))))))
	if err == nil {
		fmt.Println("Successfully queried: ", users)
	} else {
		return err
	}

	return nil
}

// Wrapper01 implements gocql.Marshaller, gocql.Unmarshaller
// it uses string representation of net.IPNet
type Wrapper01 struct {
	ip *net.IPNet
}

// MarshalCQL serializes the string representation of net.IPNet
func (w *Wrapper01) MarshalCQL(info gocql.TypeInfo) ([]byte, error) {

	if w.ip == nil {
		return []byte{}, nil
	}

	return []byte(w.ip.String()), nil
}

// UnmarshalCQL deserializes the string representation of net.IPNet
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

// String delegates to the ip.String()
func (w *Wrapper01) String() string {
	if w.ip != nil {
		return w.ip.String()
	}

	return ""
}

// Udt03 is a simple User Defined Type with two string fields
type Udt03 struct {
	Tx  string `cql:"tx"`
	Tx2 string `cql:"tx2"`
	//Inet1 string
}

func (u *Udt03) String() string {
	return "{" + u.Tx + ", " + u.Tx2 /*+ ", " + u.Inet1*/ + "}"
}

// Udt04 is a nested User Defined Type
type Udt04 struct {
	Ahoj string `cql:"ahoj"`
	Caf  *Udt03 `cql:"caf"`
	//Inet1 string
}

func (u *Udt04) String() string {
	return "{" + u.Ahoj + ", " + u.Caf.String() /*+ ", " + u.Inet1*/ + "}"
}
