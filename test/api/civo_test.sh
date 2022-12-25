#!/bin/sh

cd ../../api || echo "FAILED to change the PWD"

docker build --target civoTest -t civo .

docker run --rm civo

docker rmi -f civo