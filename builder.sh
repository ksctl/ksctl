#!/bin/bash

# Get the binary from the source code

cd ./cli || echo -e "\033[1;31mPath couldn't be found\033[0m\n"
# Check if sudo access
go get -d
go build -v -o ksctl .
chmod +x ksctl

sudo mv -v ksctl /usr/local/bin/ksctl

# Setup the configurations dir
mkdir -p ${HOME}/.ksctl/cred

mkdir -p ${HOME}/.ksctl/config/civo/ha
mkdir -p ${HOME}/.ksctl/config/azure/ha
mkdir ${HOME}/.ksctl/config/azure/managed
mkdir ${HOME}/.ksctl/config/civo/managed
mkdir ${HOME}/.ksctl/config/aws
mkdir ${HOME}/.ksctl/config/local

echo -e "\033[1;32mINSTALL COMPLETE\033[0m\n"

cd - || echo -e "\033[1;31mFailed to move to previous directory\033[0m\n"
