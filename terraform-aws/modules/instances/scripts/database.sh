#!/bin/bash
apt update
apt install -y mysql-server

mysql -e "create user 'ksctl' identified by '${random_password.k3s_db_password.result}';"
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