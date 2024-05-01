#!/bin/sh

EXEC=$1

cd ../internal/ || exit 1

echo "--------------------------------------------"
echo "|   Testing (internal/k8sdistros)"
echo "--------------------------------------------"

cd k8sdistros/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -

