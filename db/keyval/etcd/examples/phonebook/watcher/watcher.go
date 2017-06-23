package main

import (
	"fmt"
	"github.com/ligato/cn-infra/db"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/cn-infra/db/keyval/etcd/examples/phonebook/model/phonebook"
	"os"
	"os/signal"
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

func printContact(c *phonebook.Contact) {
	fmt.Printf("\t%s\n\t\t%s\n\t\t%s\n", c.Name, c.Company, c.Phonenumber)
}

func main() {

	cfg, err := processArgs()
	if err != nil {
		printUsage()
		fmt.Println(err)
		os.Exit(1)
	}

	//create connection to etcd
	db, err := etcd.NewBytesBrokerEtcd(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//initialize proto decorator
	protoBroker := etcd.NewProtoBrokerEtcd(db)

	respChan := make(chan keyval.ProtoWatchResp, 0)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	err = protoBroker.Watch(respChan, phonebook.EtcdPath())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Watching the key: ", phonebook.EtcdPath())

watcherLoop:
	for {
		select {
		case resp := <-respChan:
			switch resp.GetChangeType() {
			case data.Put:
				contact := &phonebook.Contact{}
				fmt.Println("Creating ", resp.GetKey())
				resp.GetValue(contact)
				printContact(contact)
			case data.Delete:
				fmt.Println("Removing ", resp.GetKey())
			}
			fmt.Println("============================================")
		case <-sigChan:
			break watcherLoop
		}
	}
	fmt.Println("Stop requested ...")
}
