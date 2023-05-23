#!/bin/bash


sudo rm -vf /usr/local/bin/ksctl


# Setup the ~/.kube
rm -rfv ${HOME}/.ksctl

echo -e "\033[1;32mUNINSTALL COMPLETE\033[0m\n"
