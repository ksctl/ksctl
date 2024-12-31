// Copyright 2024 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
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
	"flag"
	"fmt"
	"os"
	"strings"
)

// jq -r '.join(",")' <>

func main() {
	ips := flag.String("ips", "", "Public IP addresses for ansible inventory")
	user := flag.String("user", "root", "default ssh user")
	sshLoc := flag.String("ssh-pvt-loc", "~/.ssh/id_rsa", "ssh private key loc")
	locationInventoryFile := flag.String("inventory-loc", "inventory.ini", "ansible inventory file loc")
	flag.Parse()

	inventoryFile := strings.Builder{}
	inventoryFile.WriteString("[e2e]")
	for i, ip := range strings.Split(*ips, ",") {
		inventoryFile.WriteString(fmt.Sprintf(`
vm-%d ansible_host=%s ansible_ssh_private_key_file=%s ansible_user=%s ansible_connection=ssh`, i, ip, *sshLoc, *user))
	}

	_inventory := inventoryFile.String()

	if err := os.WriteFile(*locationInventoryFile, []byte(_inventory), 0750); err != nil {
		panic(err)
	}
	fmt.Printf("written the inventory file, %s\n", *locationInventoryFile)
}
