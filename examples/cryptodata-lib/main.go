//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package main

import (
	"fmt"
	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/db/cryptodata"
	"os"
)

func main() {
	// Parse configuration containing paths to private keys
	var cfg cryptodata.Config
	err := config.ParseConfigFromYamlFile("cryptodata.conf", &cfg)
	if err != nil {
		fmt.Printf("Error %v\n", err)
		return
	}

	// Create cryptodata client
	client, err := cryptodata.NewClient(cfg)
	if err != nil {
		fmt.Printf("Error %v\n", err)
		return
	}

	// Pass 1st argument from CLI as string to encrypt
	input := []byte(os.Args[1])
	fmt.Printf("Input %v\n", string(input))

	// Encrypt input string using public key from private key
	encrypted, err := client.EncryptArbitrary(input)
	if err != nil {
		fmt.Printf("Error %v\n", err)
		return
	}
	fmt.Printf("Encrypted %v\n", encrypted)

	// Decrypt previously encrypted input string
	decrypted, err := client.DecryptArbitrary(encrypted)
	if err != nil {
		fmt.Printf("Error %v\n", err)
		return
	}
	fmt.Printf("Decrypted %v\n", string(decrypted))
}
