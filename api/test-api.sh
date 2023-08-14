#!/bin/sh

echo "-----------------------------------"
echo "|   Testing (api/utils)"
echo "-----------------------------------"

cd provider/utils/
go test -v . -timeout 10s && cd -

echo "-----------------------------------"
echo "|   Testing (api/k8s_distro/k3s)    |"
echo "-----------------------------------"

cd k8s_distro/k3s/
go test -v . && cd -

echo "-----------------------------------"
echo "|   Testing (api/provider/local)"
echo "-----------------------------------"

cd provider/local/
go test . -v && cd -


echo "-----------------------------------"
echo "|   Testing (api/provider/civo)"
echo "-----------------------------------"

cd provider/civo/
go test . -v && cd -

echo "-----------------------------------"
echo "|   Testing (api/provider/azure)"
echo "-----------------------------------"

cd provider/azure/
go test . -v && cd -

rm -rvf ${HOME}/.ksctl
