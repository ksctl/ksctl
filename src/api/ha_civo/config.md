# this file is for DEVELOPMENT

## Database setup commands

> *Note*
> generate the password randome function


```bash
apt update
apt install -y mysql-server iptables iptables-persistent 

mysql -e "create user 'ksctl' identified by '234ed34r4345Cg4';"
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

iptables -A INPUT -p tcp --destination-port 3306 -j ACCEPT
iptables-save
# to ensure check if sudo cat /etc/iptables/rules.v4 has the contents

```

## Nginx LB

```bash
apt update
apt install nginx -y

cat <<EOF > /etc/nginx/nginx.conf
user www-data;
worker_processes auto;
pid /run/nginx.pid;
include /etc/nginx/modules-enabled/*.conf;

events {}
stream {
        upstream k3s_servers {
                server 192.168.1.9:6443;
                server 192.168.1.10:6443;
                server 192.168.1.11:6443;
        }
        server {
                listen 6443;
                proxy_pass k3s_servers;
        }
}
EOF

systemctl restart nginx
```

## HAProxy LB

```bash
apt install haproxy -y

systemctl start haproxy && systemctl enable haproxy

cat <<EOF > /etc/haproxy/haproxy.cfg
frontend kubernetes-frontend
  bind *:6443
  mode tcp
  option tcplog
  timeout client 10s
  default_backend kubernetes-backend

backend kubernetes-backend
  timeout connect 10s
  timeout server 10s
  mode tcp
  option tcp-check
  balance roundrobin
  server k3sserver1 <privateip>:6443 
  server k3sserver2 <privateip>:6443
EOF

systemctl restart haproxy
```

> **Note**
> IP address provided is the private IP for the loadblancer
> token needs to be extracted 
> database endpoint also

## Control-plane-1

> first create instances then uses the obj of loadbalancer to config it then call the control-plane-1 with first 2 lines of code
then extract the last token using the withOutput()
> controlplane-2 ,... withoutOutput()

```bash
export K3S_DATASTORE_ENDPOINT='mysql://ksctl:234ed34r4345Cg4@tcp(192.168.1.7:3306)/ksctldb'
curl -sfL https://get.k3s.io | sh -s - server --node-taint CriticalAddonsOnly=true:NoExecute --tls-san 192.168.1.8
cat /var/lib/rancher/k3s/server/token
```

## Control-plane-2 and 3 ...

```bash
export SECRET='K1083938a89a9c6176b3f9e94281f4a7203c8481a21f90d8c13874aaa00b1164a42::server:3a634d13bf04dfd17de8ff4d3ac0010c'
export K3S_DATASTORE_ENDPOINT='mysql://ksctl:234ed34r4345Cg4@tcp(192.168.1.7:3306)/ksctldb'

curl -sfL https://get.k3s.io | sh -s - server --token=$SECRET --node-taint CriticalAddonsOnly=true:NoExecute --tls-san 192.168.1.8
```

## Worker Nodes 1,2...

> After its control-plane done then worker node script using withoutscript()

```bash
export SECRET='K1083938a89a9c6176b3f9e94281f4a7203c8481a21f90d8c13874aaa00b1164a42::server:3a634d13bf04dfd17de8ff4d3ac0010c'

curl -sfL https://get.k3s.io | sh -s - agent --token=$SECRET --server https://192.168.1.8:6443

```

> **IMP** Copy the kubeconfig from any one controlplane and replace the ip with public ip of load balancer

