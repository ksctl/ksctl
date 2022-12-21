package ha_civo

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
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
func (obj *HAType) CreateDatabase() (string, error) {

	errV := obj.CreateNetwork(obj.ClusterName + "-ksctl")
	if errV != nil {
		return "", errV
	}

	name := obj.ClusterName + "-ksctl-db"

	firewall, err := obj.CreateFirewall(name)
	if err != nil {
		return "", err
	}
	obj.DBFirewallID = firewall.ID

	err = obj.ConfigWriterFirewall(firewall)
	if err != nil {
		return "", nil
	}

	instance, err := obj.CreateInstance(name, firewall.ID, "g3.large", "")
	if err != nil {
		return "", err
	}

	err = obj.ConfigWriterInstance(instance)
	if err != nil {
		return "", nil
	}

	for {
		getInstance, err := obj.GetInstance(instance.ID)
		if err != nil {
			return "", err
		}

		if getInstance.Status == "ACTIVE" {

			generatedPassword := generateDBPassword(20)
			log.Println("✅ 🚀 Instance " + name)
			err = ExecWithoutOutput(getInstance.PublicIP, getInstance.InitialPassword, scriptDB(generatedPassword), false)
			if err != nil {
				return "", err
			}

			log.Println("✅ 🔧🔨 Database")
			return fmt.Sprintf("mysql://ksctl:%s@tcp(%s:3306)/ksctldb", generatedPassword, getInstance.PrivateIP), nil
		}
		log.Println("🚧 Instance " + name)
		time.Sleep(10 * time.Second)
	}
}
