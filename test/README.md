# For the profiling Mock tests

```bash
go test -cpuprofile ../out/cpu.prof -memprofile ../out/mem.prof -bench . -benchtime=1x -cover -v
go tool pprof cpu.prof
go tool pprof mem.prof
```

# setting up github custom runner for e2e

build is done using github managed runner
run is done using github custom runner

> [!TODO]
> Need to add the instructions

```bash

```

# setting up jenkins for e2e

## jenkins

```bash
sudo apt update -y
sudo apt upgrade -y
sudo apt install openjdk-17-jre
java -version
curl -fsSL https://pkg.jenkins.io/debian-stable/jenkins.io-2023.key | sudo tee   /usr/share/keyrings/jenkins-keyring.asc > /dev/null
echo deb [signed-by=/usr/share/keyrings/jenkins-keyring.asc]   https://pkg.jenkins.io/debian-stable binary/ | sudo tee   /etc/apt/sources.list.d/jenkins.list > /dev/null
sudo apt-get update
sudo apt-get install jenkins
systemctl status jenkins
```

## Docker

```bash
sudo apt install docker.io
sudo usermod -aG docker jenkins
sudo usermod -aG docker ubuntu  # assuming current user is ubuntu
```

## Go

```bash
wget https://go.dev/dl/go1.21.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.1.linux-amd64.tar.gz
```

## Kubectl

[Link to Docs](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/#install-using-native-package-management)

## Let's Encrypt and Nginx

```bash
sudo apt-get update
sudo apt-get install certbot
sudo apt-get install python3-certbot-nginx
```

```conf
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    root /var/www/html;
    server_name XXy.Y.z;
    location / {
        proxy_pass http://127.0.0.1:8080;
    }
}
```

> Docs: https://www.nginx.com/blog/using-free-ssltls-certificates-from-lets-encrypt-with-nginx/

## Server final setup

make the user access to sudo

```bash
sudo visudo

    jenkins ALL=(ALL) NOPASSWD:ALL

sudo apt install gcc g++ -y

# first run do the install manually then call the jenkins build
as it assumes that the ksctl was there so there is rm dir command
```
