package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/consul"
	"github.com/ligato/cn-infra/db/keyval/kvproto"
	"github.com/ligato/cn-infra/examples/etcdv3-lib/model/phonebook"
)

func main() {
	db, err := consul.NewConsulStore("127.0.0.1:8500")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	protoDb := kvproto.NewProtoWrapper(db)
	defer protoDb.Close()

	put(protoDb, []string{"TheName", "TheCompany", "123456"})
	get(protoDb, []string{"TheName"})

	resp, err := protoDb.ListValues(phonebook.EtcdPath())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var revision int64
	fmt.Println("Phonebook:")
	for {
		c := &phonebook.Contact{}
		kv, stop := resp.GetNext()
		if stop {
			break
		}
		if kv.GetRevision() > revision {
			revision = kv.GetRevision()
		}
		err = kv.GetValue(c)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("\t%s\n\t\t%s\n\t\t%s\n", c.Name, c.Company, c.Phonenumber)

	}
	fmt.Println("Revision", revision)

}

func put(db keyval.ProtoBroker, data []string) {
	c := &phonebook.Contact{Name: data[0], Company: data[1], Phonenumber: data[2]}

	key := KeyContactPath(c)

	err := db.Put(key, c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Saved:", key)
}

func get(db keyval.ProtoBroker, data []string) {
	c := &phonebook.Contact{Name: data[0]}

	key := KeyContactPath(c)

	found, _, err := db.GetValue(key, c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !found {
		fmt.Println("Not found")
		return
	}

	fmt.Println("Loaded:", key, c)
}

// EtcdPath returns the base path were the phonebook records are stored.
func KeyPath() string {
	return "phonebook/"
}

// EtcdContactPath returns the path for a given contact.
func KeyContactPath(contact *phonebook.Contact) string {
	return KeyPath() + strings.Replace(contact.Name, " ", "", -1)
}
