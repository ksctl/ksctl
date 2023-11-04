#!/bin/sh

cd ../pkg/ || exit 1

echo "-----------------------------------"
echo "|   Testing (pkg/utils)"
echo "-----------------------------------"

cd utils/
go test -fuzz=Fuzz -fuzztime 10s -v cloud_test.go utils.go
go test -fuzz=Fuzz -fuzztime 10s -v cni_test.go utils.go
go test -fuzz=Fuzz -fuzztime 10s -v name_test.go utils.go
go test -fuzz=Fuzz -fuzztime 10s -v storage_test.go utils.go
go test -fuzz=Fuzz -fuzztime 10s -v distro_test.go utils.go
go test utils_test.go utils.go -v && cd -
