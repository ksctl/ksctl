#!/bin/sh

EXEC=$1
cd ../ || exit 1

echo "--------------------------------------------"
echo "|   Testing (poller/)"
echo "--------------------------------------------"

cd poller
GOTEST_PALETTE="red,yellow,green" $EXEC ./... -v && cd -
