#!/bin/sh

cd ../internal/ || exit 1

echo "--------------------------------------------"
echo "|   Testing (internal/cloudproviders/local)"
echo "--------------------------------------------"

cd cloudproviders/local/
go test . -v && cd -
