#!/bin/sh

EXEC=$1
cd ../pkg/ || exit 1

echo "-----------------------------------"
echo "|   Testing (pkg/helpers)"
echo "-----------------------------------"

cd helpers/
GOTEST_PALETTE="red,yellow,green" $EXEC -fuzz=Fuzz -fuzztime 10s -v cloud_test.go fields.go
GOTEST_PALETTE="red,yellow,green" $EXEC -fuzz=Fuzz -fuzztime 10s -v cni_test.go fields.go
GOTEST_PALETTE="red,yellow,green" $EXEC -fuzz=Fuzz -fuzztime 10s -v name_test.go fields.go
GOTEST_PALETTE="red,yellow,green" $EXEC -fuzz=Fuzz -fuzztime 10s -v storage_test.go fields.go
GOTEST_PALETTE="red,yellow,green" $EXEC -fuzz=Fuzz -fuzztime 10s -v distro_test.go fields.go
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -
