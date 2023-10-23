#!/bin/sh

cd ../internal

echo "--------------------------------------------"
echo "|   Testing (internal/cloudproviders/azure)"
echo "--------------------------------------------"

cd cloudproviders/azure/
go test . -v && cd -

