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

> Docs: https://www.nginx.com/blog/using-free-ssltls-certificates-from-lets-encrypt-with-nginx/
