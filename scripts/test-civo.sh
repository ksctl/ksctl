#!/bin/sh
EXEC=$1
cd ../internal || exit 1

echo "--------------------------------------------"
echo "|   Testing (internal/cloudproviders/civo)"
echo "--------------------------------------------"

cd cloudproviders/civo/
GOTEST_PALETTE="red,yellow,green" $EXEC -tags testing_civo ./... -v && cd -

