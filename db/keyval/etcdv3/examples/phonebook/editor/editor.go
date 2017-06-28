package main

import (
	"encoding/json"
	"fmt"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/db/keyval/etcdv3/examples/phonebook/model/phonebook"
	"os"
)

const (
	// Put represents put operation of single key-value pair
	Put = iota
	// PutTxn represents put operation used in transaction
	PutTxn = iota
	// Delete represents delete operation
	Delete = iota
)

func processArgs() (cfg string, op int, data []string, err error) {
	var task []string

	//default args
	cfg = ""
	op = Put

	if len(os.Args) > 2 {
		if os.Args[1] == "--cfg" {
			cfg = os.Args[2]
			task = os.Args[3:]
		} else {
			task = os.Args[1:]
		}
	} else {
		return "", 0, nil, fmt.Errorf("Incorrect arguments.")
	}

	if len(task) < 2 || (task[0] == "put" && len(task) < 4) {
		return "", 0, nil, fmt.Errorf("Incorrect arguments.")
	}

	if task[0] == "delete" {
		op = Delete
	} else if task[0] == "puttxn" {
		op = PutTxn
	}

	return cfg, op, task[1:], nil
}

func printUsage() {
	fmt.Printf("\n\n%s: [--cfg CONFIG_FILE] <delete NAME | put NAME COMPANY PHONE | puttxn JSONENCODED_CONTACTS>\n\n", os.Args[0])
}

func put(db keyval.ProtoBroker, data []string) {
	c := &phonebook.Contact{Name: data[0], Company: data[1], Phonenumber: data[2]}

	key := phonebook.EtcdContactPath(c)

	//Insert the key-value pair
	db.Put(key, c)

	fmt.Println("Saving ", key)
}

func putTxn(db keyval.ProtoBroker, data string) {
	contacts := []phonebook.Contact{}

	json.Unmarshal([]byte(data), &contacts)

	txn := db.NewTxn()

	for i := range contacts {

		key := phonebook.EtcdContactPath(&contacts[i])
		fmt.Println("Saving ", key)
		//add the key-value pair into transaction
		txn.Put(key, &contacts[i])
	}

	txn.Commit()

}

func delete(db keyval.ProtoBroker, name string) {
	key := phonebook.EtcdContactPath(&phonebook.Contact{Name: name})

	//Remove the key
	db.Delete(key)
	fmt.Println("Removing ", key)
}

func main() {

	cfg, op, data, err := processArgs()
	if err != nil {
		printUsage()
		fmt.Println(err)
		os.Exit(1)
	}

	db, err := etcdv3.NewEtcdConnectionWithBytes(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//initialize proto decorator
	protoDb := etcdv3.NewProtoWrapperEtcd(db)

	switch op {
	case Put:
		put(protoDb, data)
	case PutTxn:
		putTxn(protoDb, data[0])
	case Delete:
		delete(protoDb, data[0])
	default:
		fmt.Println("Unknown operation")
	}

}
