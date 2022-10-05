#!/bin/sh

# Get the binary from the source code

usr=$(whoami)

#cd /home/$usr

#echo $(pwd)

#wget -O ~URL_OF_BINARY~

# Check if sudo access


sudo rm -vf /usr/local/bin/kubesimpctl

# Setup the configurations dir
sudo rm -rfv /etc/kubesimpctl

# Setup the ~/.kube
rm -rfv /home/$usr/.kube/kubesimpctl

echo "UNINSTALL COMPLETE"