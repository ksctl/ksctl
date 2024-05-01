#!/bin/sh

EXEC=$1
cd ../internal/ || exit 1

echo "--------------------------------------------"
echo "|   Testing (internal/kubernetes)"
echo "--------------------------------------------"

cd kubernetes
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -

