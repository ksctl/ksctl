#!/bin/bash

# Get the binary from the source code

usr=$(whoami)

#wget -O ~URL_OF_BINARY~

sudo rm -vf /usr/local/bin/ksctl

# Setup the configurations dir
#sudo rm -rfv /etc/ksctl

# Setup the ~/.kube
rm -rfv /home/$usr/.kube/ksctl

echo -e "\033[32;40mUNINSTALL COMPLETE\033[0m\n"