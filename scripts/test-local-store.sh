#!/bin/sh

EXEC=$1

cd ../internal || exit 1

echo "-----------------------------------"
echo "|   Testing (internal/storage/local)"
echo "-----------------------------------"

cd storage/local/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -
