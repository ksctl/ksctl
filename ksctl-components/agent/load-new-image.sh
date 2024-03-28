#!/bin/bash

PS4='+\[\033[0;33m\](\[\033[0;36m\]${BASH_SOURCE##*/}:${LINENO}\[\033[0;33m\])\[\033[0m\] '

set -xe

cd ../..

docker build --file build/agent/Dockerfile --tag ghcr.io/ksctl/ksctl-agent:latest .

sudo docker push ghcr.io/ksctl/ksctl-agent:latest

#kind create cluster --name test

#kind load docker-image ghcr.io/ksctl/ksctl-agent:latest --name test
