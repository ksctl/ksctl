#!/bin/sh

cd ../internal/ || exit 1

echo "--------------------------------------------"
echo "|   Testing (internal/k8sdistros)"
echo "--------------------------------------------"

cd k8sdistros/
go test . -v && cd -

