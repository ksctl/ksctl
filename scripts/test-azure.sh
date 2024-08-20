#!/bin/sh

EXEC=$1

cd ../internal

echo "--------------------------------------------"
echo "|   Testing (internal/cloudproviders/azure)"
echo "--------------------------------------------"

cd cloudproviders/azure/
GOTEST_PALETTE="red,yellow,green" $EXEC -tags testing_azure -v ./... && cd -

