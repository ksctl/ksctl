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
