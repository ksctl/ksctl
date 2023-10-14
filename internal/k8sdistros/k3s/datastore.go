package k3s

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/kubesimplify/ksctl/pkg/resources"
	"github.com/kubesimplify/ksctl/pkg/utils"

	. "github.com/kubesimplify/ksctl/pkg/utils/consts"
)

// ConfigureDataStore implements resources.DistroFactory.
func (k3s *K3sDistro) ConfigureDataStore(idx int, storage resources.StorageFactory) error {
	storage.Logger().Print("[k3s] configuring Datastore", strconv.Itoa(idx))

	if idx > 0 {
		storage.Logger().Note("[k3s] cluster of datastore not enabled!", strconv.Itoa(idx))
		return nil
	}

	password := generateDBPassword(15)

	err := k3s.SSHInfo.Flag(UtilExecWithoutOutput).Script(
		scriptDB(password)).
		IPv4(k8sState.PublicIPs.DataStores[idx]).
		FastMode(true).SSHExecute(storage)
	if err != nil {
		return err
	}
	k8sState.DataStoreEndPoint = fmt.Sprintf("mysql://ksctl:%s@tcp(%s:3306)/ksctldb", password, k8sState.PrivateIPs.DataStores[idx])

	path := utils.GetPath(UtilClusterPath, k8sState.Provider, k8sState.ClusterType, k8sState.ClusterDir, STATE_FILE_NAME)
	err = saveStateHelper(storage, path)
	if err != nil {
		return err
	}
	storage.Logger().Success("[k3s] configured DataStore", strconv.Itoa(idx))

	return nil
}

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
