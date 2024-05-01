#!/bin/sh

EXEC=$1

cd ../internal || exit 1

echo "-----------------------------------"
echo "|   Testing (internal/storage/kubernetes)"
echo "-----------------------------------"

cd storage/kubernetes/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -
