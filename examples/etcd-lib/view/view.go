//go:generate protoc --proto_path=../model/phonebook --go_out=../model/phonebook ../model/phonebook/phonebook.proto

// Package view contains an example that shows how to read data from etcd.
package main

import (
	"fmt"
	"os"

	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/db/keyval/etcd"
	"go.ligato.io/cn-infra/v2/db/keyval/kvproto"
	"go.ligato.io/cn-infra/v2/examples/etcd-lib/model/phonebook"
	"go.ligato.io/cn-infra/v2/logging/logs"
)

// processArgs processes input arguments.
func processArgs() (*etcd.ClientConfig, error) {
	fileConfig := &etcd.Config{}
	if len(os.Args) > 2 {
		if os.Args[1] == "--cfg" {

			err := config.ParseConfigFromYamlFile(os.Args[2], fileConfig)
			if err != nil {
				return nil, err
			}

		} else {
			return nil, fmt.Errorf("incorrect arguments")
		}
	}

	return etcd.ConfigToClient(fileConfig)
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

	// Create connection to etcd.
	db, err := etcd.NewEtcdConnectionWithBytes(*cfg, logs.DefaultLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Initialize proto decorator.
	protoDb := kvproto.NewProtoWrapper(db)

	// Retrieve all contacts from database.
	resp, err := protoDb.ListValues(phonebook.EtcdPath())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Print out all contacts one-by-one.
	var revision int64
	fmt.Println("Phonebook:")
	for {
		c := &phonebook.Contact{}
		kv, stop := resp.GetNext()
		if stop {
			break
		}
		// Maintain the latest revision.
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
