#!/bin/sh

cd ../internal || exit 1

echo "-----------------------------------"
echo "|   Testing (internal/storage/kubernetes)"
echo "-----------------------------------"

cd storage/kubernetes/
go test . -v && cd -
