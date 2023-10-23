#!/bin/sh

cd ../pkg/ || exit 1

echo "-----------------------------------"
echo "|   Testing (pkg/logger)"
echo "-----------------------------------"

cd logger/
go test . -v -timeout 10s && cd -

