#!/bin/sh

cd ../internal || exit 1

echo "-----------------------------------"
echo "|   Testing (internal/storage/local)"
echo "-----------------------------------"

cd storage/local/
go test . -v && cd -
