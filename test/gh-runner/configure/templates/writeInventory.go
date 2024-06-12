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
