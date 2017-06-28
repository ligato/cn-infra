//go:generate protoc --proto_path=../model/phonebook --gogo_out=../model/phonebook ../model/phonebook/phonebook.proto
package main

import (
	"fmt"
	"os"

	"github.com/coreos/etcd/clientv3"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/db/keyval/etcdv3/examples/phonebook/model/phonebook"
	"github.com/ligato/cn-infra/utils/config"
)

func processArgs() (*clientv3.Config, error) {
	fileConfig := &etcdv3.Config{}
	if len(os.Args) > 2 {
		if os.Args[1] == "--cfg" {

			err := config.ParseConfigFromYamlFile(os.Args[2], fileConfig)
			if err != nil {
				return nil, err
			}

		} else {
			return nil, fmt.Errorf("Incorrect arguments.")
		}
	}

	return etcdv3.ConfigToClientv3(fileConfig)
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
	db, err := etcdv3.NewEtcdConnectionWithBytes(*cfg)
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
	protoDb.Close()
}
