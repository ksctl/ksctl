#!/bin/bash

# Get the binary from the source code

#wget -O ~URL_OF_BINARY~

cd src/cli || echo -e "\033[31;40mPath couldn't be found\033[0m\n"
# Check if sudo access
go get -d
go build -v -o ksctl .
cd - || echo -e "\033[31;40mFailed to move to previous directory\033[0m\n"
chmod +x src/cli/ksctl


sudo mv -v src/cli/ksctl /usr/local/bin/ksctl
#cp -v ./ksctl /usr/local/bin

# Setup the configurations dir
#sudo mkdir /etc/ksctl

# Setup the ~/.ksctl
mkdir -p ${HOME}/.ksctl/cred
touch ${HOME}/.ksctl/cred/aws
touch ${HOME}/.ksctl/cred/azure
touch ${HOME}/.ksctl/cred/civo

mkdir -p ${HOME}/.ksctl/config/civo
mkdir ${HOME}/.ksctl/config/azure
mkdir ${HOME}/.ksctl/config/aws
mkdir ${HOME}/.ksctl/config/local

echo -e "\033[32;40mINSTALL COMPLETE\033[0m\n"