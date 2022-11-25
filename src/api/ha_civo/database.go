package ha_civo

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/civo/civogo"
)

// TODO: perform cleanup when there is error

func generateDBPassword(passwordLen int) string {
	var password strings.Builder
	var (
		lowerCharSet = "abcdedfghijklmnopqrst"
		upperCharSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numberSet    = "0123456789"
		allCharSet   = lowerCharSet + upperCharSet + numberSet
	)
	rand.Seed(time.Now().Unix())

	for i := 0; i < passwordLen; i++ {
		random := rand.Intn(len(allCharSet))
		password.WriteString(string(allCharSet[random]))
	}

	inRune := []rune(password.String())
	rand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})

	return string(inRune)
}

func scriptDB(password string) string {
	return fmt.Sprintf(`#!/bin/bash
apt update
apt install -y mysql-server

systemctl start mysql && systemctl enable mysql

mysql -e "create user 'ksctl' identified by '%s';"
mysql -e "create database ksctldb; grant all on ksctldb.* to 'ksctl';"

cat <<EOF > /etc/mysql/mysql.conf.d/mysqld.cnf
[mysqld]
user            = mysql
pid-file        = /var/run/mysqld/mysqld.pid
socket  = /var/run/mysqld/mysqld.sock
port            = 3306
datadir = /var/lib/mysql

bind-address            = 0.0.0.0
mysqlx-bind-address     = 0.0.0.0
key_buffer_size         = 16M

myisam-recover-options  = BACKUP

log_error = /var/log/mysql/error.log
max_binlog_size   = 100M

EOF

systemctl restart mysql
`, password)
}

// CreateDatabase return endpoint address if no error is encountered
func CreateDatabase(client *civogo.Client, clusterName string) (string, error) {

	var networkID string

	network, err := CreateNetwork(client, clusterName+"-ksctl")
	if err != nil {
		return "", err
	}

	err = ConfigWriterNetwork(network, clusterName, client.Region)
	if err != nil {
		return "", err
	}

	networkID = network.ID

	diskImg, err := client.GetDiskImageByName("ubuntu-focal")
	if err != nil {
		return "", err
	}

	firewall, err := CreateFirewall(client, clusterName+"-ksctl-db", networkID)
	if err != nil {
		return "", err
	}

	err = ConfigWriterFirewall(firewall, clusterName, client.Region)
	if err != nil {
		return "", nil
	}

	instance, err := CreateInstance(client, clusterName+"-ksctl-db", firewall.ID, diskImg.ID, "g3.large", networkID, "")
	if err != nil {
		return "", err
	}

	err = ConfigWriterInstance(instance, clusterName, client.Region)
	if err != nil {
		return "", nil
	}

	for {
		getInstance, err := GetInstance(client, instance.ID)
		if err != nil {
			return "", err
		}

		if getInstance.Status == "ACTIVE" {
			generatedPassword := generateDBPassword(20)
			log.Println("[ CREATED ] Instance " + clusterName + "-ksctl-db")
			err = ExecWithoutOutput(getInstance.PublicIP, getInstance.InitialPassword, scriptDB(generatedPassword), false)
			if err != nil {
				return "", err
			}

			log.Println("[CONFIGURED] Database")
			return fmt.Sprintf("mysql://ksctl:%s@tcp(%s:3306)/ksctldb", generatedPassword, getInstance.PrivateIP), nil
		}
		log.Println(getInstance.Status)
		time.Sleep(10 * time.Second)
	}
}
