#!/bin/sh
EXEC=$1
cd ../internal || exit 1

echo "--------------------------------------------"
echo "|   Testing (internal/cloudproviders/aws)"
echo "--------------------------------------------"

cd cloudproviders/aws/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -

