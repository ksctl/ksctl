#!/bin/bash

# Get the binary from the source code

usr=$(whoami)

#wget -O ~URL_OF_BINARY~

cd src/cli || echo -e "\033[31;40mPath couldn't be found\033[0m\n"
# Check if sudo access
go build -v -o kubesimpctl .
cd - || echo -e "\033[31;40mFailed to move to previous directory\033[0m\n"
chmod +x src/cli/kubesimpctl


sudo mv -v src/cli/kubesimpctl /usr/local/bin/kubesimpctl
#cp -v ./kubesimpctl /usr/local/bin

# Setup the configurations dir
#sudo mkdir /etc/kubesimpctl

# Setup the ~/.kube
mkdir -p /home/$usr/.kube/kubesimpctl/cred
touch /home/$usr/.kube/kubesimpctl/cred/aws
touch /home/$usr/.kube/kubesimpctl/cred/azure
touch /home/$usr/.kube/kubesimpctl/cred/civo

mkdir -p /home/$usr/.kube/kubesimpctl/config/civo
mkdir /home/$usr/.kube/kubesimpctl/config/azure
mkdir /home/$usr/.kube/kubesimpctl/config/aws
mkdir /home/$usr/.kube/kubesimpctl/config/local

echo -e "\033[32;40mINSTALL COMPLETE\033[0m\n"