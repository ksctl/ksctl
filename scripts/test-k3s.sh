#!/bin/sh

cd ../internal/ || exit 1

echo "--------------------------------------------"
echo "|   Testing (internal/k8sdistros/k3s)"
echo "--------------------------------------------"

cd k8sdistros/k3s/
go test . -v && cd -

