#!/bin/sh

EXEC=$1
cd ../internal/ || exit 1

echo "--------------------------------------------"
echo "|   Testing (internal/k8sdistros/k3s)"
echo "--------------------------------------------"

cd k8sdistros/k3s/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -

