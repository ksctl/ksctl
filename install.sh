#!/bin/bash

# Get the binary from the source code

usr=$(whoami)

#cd /home/$usr

#echo $(pwd)

#wget -O ~URL_OF_BINARY~

cd src/cli || echo "Path couldn't be found"
# Check if sudo access
go build -v -o kubesimpctl .
cd - || echo "Failed to move to previous directory"
chmod +x src/cli/kubesimpctl


sudo cp -v src/cli/kubesimpctl /usr/local/bin/kubesimpctl
#cp -v ./kubesimpctl /usr/local/bin

# Setup the configurations dir
sudo mkdir /etc/kubesimpctl

# Setup the ~/.kube
mkdir -p /home/$usr/.kube/kubesimpctl/config
mkdir -p /home/$usr/.kube/kubesimpctl/cred
touch /home/$usr/.kube/kubesimpctl/cred/aws
touch /home/$usr/.kube/kubesimpctl/cred/azure
touch /home/$usr/.kube/kubesimpctl/cred/civo

mkdir /home/$usr/.kube/kubesimpctl/config/azure
mkdir /home/$usr/.kube/kubesimpctl/config/aws
mkdir /home/$usr/.kube/kubesimpctl/config/local

echo "INSTALL COMPLETE"