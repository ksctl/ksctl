#!/bin/sh

EXEC=$1

cd ../pkg/ || exit 1

echo "-----------------------------------"
echo "|   Testing (pkg/logger)"
echo "-----------------------------------"

cd logger/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v -timeout 10s && cd -

