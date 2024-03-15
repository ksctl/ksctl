package k3s

import (
	"fmt"
	"strconv"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/resources"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

// ConfigureDataStore implements resources.DistroFactory.
// TODO: Update the k3s state struct and also script to use cat <<EOF for passing the certs as well
func (k3s *K3sDistro) ConfigureDataStore(idx int, storage resources.StorageFactory) error {
	log.Print("configuring Datastore", "number", strconv.Itoa(idx))

	//if idx > 0 {
	//	log.Warn("cluster of datastore not enabled!", "number", strconv.Itoa(idx))
	//	return nil
	//}

	password, err := helpers.GenRandomString(15)
	if err != nil {
		return log.NewError("Error in generating random string", "reason", err.Error())
	}

	err = k3s.SSHInfo.Flag(consts.UtilExecWithoutOutput).Script(
		scriptDB(password)).
		IPv4(mainStateDocument.K8sBootstrap.K3s.B.PublicIPs.DataStores[idx]).
		FastMode(true).SSHExecute(log)
	if err != nil {
		return log.NewError(err.Error())
	}
	mainStateDocument.K8sBootstrap.K3s.DataStoreEndPoint = fmt.Sprintf("mysql://ksctl:%s@tcp(%s:3306)/ksctldb", password, mainStateDocument.K8sBootstrap.K3s.B.PrivateIPs.DataStores[idx])
	log.Debug("Printing", "datastoreEndpoint", mainStateDocument.K8sBootstrap.K3s.DataStoreEndPoint)

	err = storage.Write(mainStateDocument)
	if err != nil {
		return log.NewError(err.Error())
	}
	log.Success("configured DataStore", "number", strconv.Itoa(idx))

	return nil
}

func scriptDB(password string) string {
	return fmt.Sprintf(`#!/bin/bash
sudo apt update
sudo apt install -y mysql-server

sudo systemctl start mysql

sudo systemctl enable mysql

cat <<EOF > mysqld.cnf
[mysqld]
user		= mysql
bind-address		= 0.0.0.0
#mysqlx-bind-address	= 127.0.0.1

key_buffer_size		= 16M
myisam-recover-options  = BACKUP
log_error = /var/log/mysql/error.log
max_binlog_size   = 100M

EOF

sudo mv mysqld.cnf /etc/mysql/mysql.conf.d/mysqld.cnf

sudo systemctl restart mysql

sudo mysql -e "create user 'ksctl' identified by '%s';"
sudo mysql -e "create database ksctldb; grant all on ksctldb.* to 'ksctl';"

`, password)
}
