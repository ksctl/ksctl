#!/bin/sh

cd ../internal || exit 1

echo "--------------------------------------------"
echo "|   Testing (internal/cloudproviders/civo)"
echo "--------------------------------------------"

cd cloudproviders/civo/
go test . -v && cd -

