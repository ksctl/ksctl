#!/bin/sh

EXEC=$1

cd ../internal || exit 1

echo "-----------------------------------"
echo "|   Testing (internal/storage/external/mongodb)"
echo "-----------------------------------"

cd storage/external/mongodb/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -
