#!/bin/sh

cd ../internal || exit 1

echo "-----------------------------------"
echo "|   Testing (internal/storage/external/mongodb)"
echo "-----------------------------------"

cd storage/external/mongodb/
go test . -v && cd -
