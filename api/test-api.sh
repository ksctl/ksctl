#!/bin/sh

echo "+-------------------------+"
echo "|   Testing (api/util)    |"
echo "+-------------------------+"

cd utils/
go test . -v && cd -



echo "+-------------------------+"
echo "|   Testing (api/local)   |"
echo "+-------------------------+"

cd local/
go test . -v && cd -



echo "+-------------------------+"
echo "|   Testing (api/civo)    |"
echo "+-------------------------+"

cd civo/
go test . -v && cd -



echo "+-------------------------+"
echo "|  Testing (api/ha_civo)  |"
echo "+-------------------------+"

cd ha_civo/
go test . -v && cd -
