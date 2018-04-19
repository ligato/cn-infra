package main

import (
	"fmt"
	"log"

	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/consul"
	"github.com/ligato/cn-infra/db/keyval/kvproto"
	"github.com/ligato/cn-infra/examples/etcdv3-lib/model/phonebook"
)

func main() {
	db, err := consul.NewConsulStore("127.0.0.1:8500")
	if err != nil {
		log.Fatal(err)
	}

	protoDb := kvproto.NewProtoWrapper(db)
	defer protoDb.Close()

	list(protoDb)
	put(protoDb, []string{"TheName", "TheCompany", "123456"})
	get(protoDb, "TheName")
	list(protoDb)
	del(protoDb, "TheName")
	list(protoDb)

}

func list(db keyval.ProtoBroker) {
	resp, err := db.ListValues(phonebook.EtcdPath())
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
		}

		fmt.Printf("\t%s\n\t\t%s\n\t\t%s\n", c.Name, c.Company, c.Phonenumber)

	}
	fmt.Println("Revision", revision)
}

func put(db keyval.ProtoBroker, data []string) {
	c := &phonebook.Contact{Name: data[0], Company: data[1], Phonenumber: data[2]}

	key := phonebook.EtcdContactPath(c)

	err := db.Put(key, c)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Saved:", key)
}

func get(db keyval.ProtoBroker, data string) {
	c := &phonebook.Contact{Name: data}

	key := phonebook.EtcdContactPath(c)

	found, _, err := db.GetValue(key, c)
	if err != nil {
		log.Fatal(err)
	} else if !found {
		fmt.Println("Not found")
		return
	}

	fmt.Println("Loaded:", key, c)
}

func del(db keyval.ProtoBroker, data string) {
	c := &phonebook.Contact{Name: data}

	key := phonebook.EtcdContactPath(c)

	existed, err := db.Delete(key)
	if err != nil {
		log.Fatal(err)
	} else if !existed {
		fmt.Println("Not existed")
		return
	}

	fmt.Println("Deleted:", key)
}
