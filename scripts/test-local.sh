#!/bin/sh

EXEC=$1

cd ../internal/ || exit 1

echo "--------------------------------------------"
echo "|   Testing (internal/cloudproviders/local)"
echo "--------------------------------------------"

cd cloudproviders/local/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -
