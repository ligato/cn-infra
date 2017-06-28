//go:generate protoc --proto_path=../model/phonebook --gogo_out=../model/phonebook ../model/phonebook/phonebook.proto
package main

import (
	"fmt"
	"os"

	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/db/keyval/etcdv3/examples/phonebook/model/phonebook"
)

func processArgs() (string, error) {
	cfg := ""
	if len(os.Args) > 1 {
		if os.Args[1] == "--cfg" {
			cfg = os.Args[2]

		} else {
			return "", fmt.Errorf("Incorrect arguments.")
		}
	}

	return cfg, nil
}

func printUsage() {
	fmt.Printf("\n\n%s: [--cfg CONFIG_FILE] <delete NAME | put NAME COMPANY PHONE>\n\n", os.Args[0])
}

func main() {

	cfg, err := processArgs()
	if err != nil {
		printUsage()
		fmt.Println(err)
		os.Exit(1)
	}

	//create connection to etcd
	db, err := etcdv3.NewEtcdConnectionWithBytes(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//initialize proto decorator
	protoDb := etcdv3.NewProtoWrapperEtcd(db)

	//retrieve all contacts
	resp, err := protoDb.ListValues(phonebook.EtcdPath())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var revision int64
	fmt.Println("Phonebook:")
	for {
		c := &phonebook.Contact{}
		kv, allReceived := resp.GetNext()
		if allReceived {
			break
		}
		//maintain latest revision
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
