#!/bin/sh

cd ../../src/api || echo "FAILED to change the PWD"

docker build --target localTest -t local .

docker run --rm local

docker rmi -f local