#!/bin/bash
Red='\033[1;31m'
Green='\033[1;32m'
Blue='\033[1;34m'
Yellow='\033[1;33m'
NoColor='\033[0m' 

if command -v jq > /dev/null; then
    if command -v curl > /dev/null; then
        echo -e "${Green}Installed dependencies are present${NoColor}"
    else
        if command -v wget > /dev/null; then
            echo -e "${Green}Installed dependencies are present${NoColor}"
        else
            echo -e "${Red}dependency 'wget' not found${NoColor}"
            exit 1
        fi
        echo -e "${Red}dependency 'curl' not found${NoColor}"
        exit 1
    fi
else
    echo -e "${Red}dependency 'jq' not found${NoColor}"
    exit 1
fi

# get the release version
echo -e "${Blue}Available Releases${NoColor}"

response=$(curl --silent "https://api.github.com/repos/kubesimplify/ksctl/releases")

# Loop through the releases and extract the tag names
for release in $(echo "${response}" | jq -r '.[].tag_name'); do
    echo -e "${Blue}${release}${NoColor}"
done



echo -e "${Yellow}Enter the ksctl version to install${NoColor}"
read RELEASE_VERSION

len=$(echo "${#RELEASE_VERSION}")

RELEASE_VERSION="${RELEASE_VERSION:1:$len}"

os_name=$(uname)
OS=""
ARCH=""
arch=$(uname -m)


if [[ "$arch" == "x86_64" ]]; then
    ARCH="amd64"
elif [[ "$arch" == "arm64" ]]; then
    ARCH="arm64"
else
    echo -e "${Red}Invalid architecture${NoColor}"
    exit 1
fi

if [[ "$os_name" == "Linux" ]]; then
    OS="linux"
elif [[ "$os_name" == "Darwin" ]]; then
    OS="darwin"
else
    echo -e "${Red}Invalid OS${NoColor}"
    exit 1
fi

echo "Detected ${OS} which is ${ARCH}"

if command -v sha256sum > /dev/null; then
    echo -e "${Green}Installed dependencies are present${NoColor}"
else
    echo -e "${Red}dependency 'sha256sum' not found${NoColor}"
    echo -e "${Blue}if you are using mac you can use 'brew install coreutils'${NoColor}"
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
mkdir -p ${HOME}/.ksctl/config/aws/managed
mkdir -p ${HOME}/.ksctl/config/aws/ha
mkdir -p ${HOME}/.ksctl/config/local/managed

echo -e "${Green}INSTALL COMPLETE${NoColor}"
