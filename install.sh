#!/bin/bash
Red='\033[0;31m'
Green='\033[0;32m'
Blue='\033[0;34m'
Yellow='\033[0;33m'
NoColor='\033[0m' 


# get the release version
echo -e "${Blue}Available Releases${NoColor}"

response=$(curl --silent "https://api.github.com/repos/kubesimplify/ksctl/releases")

# Loop through the releases and extract the tag names
for release in $(echo "${response}" | jq -r '.[].tag_name'); do
    echo -e "${Blue}${release}${NoColor}"
done



echo -e "${Yellow}Enter the ksctl version to install (enter the string after 'v')${NoColor}"
read RELEASE_VERSION

echo -e "${Yellow}Enter the OS and corresponding Architecture${NoColor}"
echo -e "${Blue}Enter [1] for Linux and [0] for MacOS${NoColor}"
read OS

echo -e "${Blue}Enter [1] for amd64 or x86_64 and [0] for arm64${NoColor}"
read ARCH


if [[ $ARCH -eq 1 ]]; then
  ARCH="amd64"
elif [[ $ARCH -eq 0 ]]; then
  ARCH="arm64"
else
  echo -e "${Red}Invalid architecture${NoColor}"
  exit 1
fi

if [[ $OS -eq 1 ]]; then
  OS="linux"
elif [[ $OS -eq 0 ]]; then
  OS="darwin"
else
  echo -e "${Red}Invalid OS${NoColor}"
  exit 1
fi


cd /tmp
sudo wget -q https://github.com/kubesimplify/ksctl/releases/download/v${RELEASE_VERSION}/ksctl_${RELEASE_VERSION}_checksums.txt
sudo wget https://github.com/kubesimplify/ksctl/releases/download/v${RELEASE_VERSION}/ksctl_${RELEASE_VERSION}_${OS}_${ARCH}.tar.gz
sudo wget -q https://github.com/kubesimplify/ksctl/releases/download/v${RELEASE_VERSION}/ksctl_${RELEASE_VERSION}_${OS}_${ARCH}.tar.gz.cert

file=$(sha256sum ksctl_${RELEASE_VERSION}_${OS}_${ARCH}.tar.gz | awk '{print $1}')
checksum=$(cat ksctl_${RELEASE_VERSION}_checksums.txt | grep ksctl_${RELEASE_VERSION}_${OS}_${ARCH}.tar.gz | awk '{print $1}')

if [[ $file != $checksum ]]; then
  echo -e "${Red}Checksum didn't matched!${NoColor}"
  exit 1
else
  echo -e "${Green}CheckSum are verified${NoColor}"
fi

sudo tar -xvf ksctl_${RELEASE_VERSION}_${OS}_${ARCH}.tar.gz

sudo mv -v ksctl /usr/local/bin/ksctl
# Setup the configurations dir
mkdir -p ${HOME}/.ksctl/cred

mkdir -p ${HOME}/.ksctl/config/civo/ha
mkdir -p ${HOME}/.ksctl/config/civo/managed
mkdir -p ${HOME}/.ksctl/config/azure/ha
mkdir -p ${HOME}/.ksctl/config/azure/managed
mkdir ${HOME}/.ksctl/config/aws
mkdir ${HOME}/.ksctl/config/local

echo -e "${Green}INSTALL COMPLETE${NoColor}"
