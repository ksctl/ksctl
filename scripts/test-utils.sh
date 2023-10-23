#!/bin/sh

cd ../pkg/ || exit 1

echo "-----------------------------------"
echo "|   Testing (pkg/utils)"
echo "-----------------------------------"

cd utils/
go test -fuzz=Fuzz -fuzztime 10s -v cloud_test.go main.go
go test -fuzz=Fuzz -fuzztime 10s -v cni_test.go main.go
go test -fuzz=Fuzz -fuzztime 10s -v name_test.go main.go
go test -fuzz=Fuzz -fuzztime 10s -v storage_test.go main.go
go test -fuzz=Fuzz -fuzztime 10s -v distro_test.go main.go
go test utils_test.go main.go -v && cd -
