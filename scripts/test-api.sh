#!/bin/sh

cd ../pkg/ || exit 1

echo "-----------------------------------"
echo "|   Testing (pkg/helpers)"
echo "-----------------------------------"

cd helpers/
go test -fuzz=Fuzz -fuzztime 10s -v cloud_test.go fields.go
go test -fuzz=Fuzz -fuzztime 10s -v cni_test.go fields.go
go test -fuzz=Fuzz -fuzztime 10s -v name_test.go fields.go
go test -fuzz=Fuzz -fuzztime 10s -v storage_test.go fields.go
go test -fuzz=Fuzz -fuzztime 10s -v distro_test.go fields.go
go test . -v && cd -

echo "-----------------------------------"
echo "|   Testing (pkg/logger)"
echo "-----------------------------------"

cd logger/
go test . -v -timeout 10s && cd -

cd ../internal

echo "--------------------------------------------"
echo "|   Testing (internal/k8sdistros/k3s)"
echo "--------------------------------------------"

cd k8sdistros/k3s/
go test . -v && cd -

echo "--------------------------------------------"
echo "|   Testing (internal/cloudproviders/local)"
echo "--------------------------------------------"

cd cloudproviders/local/
go test . -v && cd -


echo "--------------------------------------------"
echo "|   Testing (internal/cloudproviders/civo)"
echo "--------------------------------------------"

cd cloudproviders/civo/
go test . -v && cd -

echo "--------------------------------------------"
echo "|   Testing (internal/cloudproviders/azure)"
echo "--------------------------------------------"

cd cloudproviders/azure/
go test . -v && cd -

