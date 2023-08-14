#!/bin/sh

echo "+-------------------------+"
echo "|   Testing (api/util)    |"
echo "+-------------------------+"

echo "TODO"

# cd utils/
# go test -v . -timeout 10s && cd -

echo "+-------------------------+"
echo "|   Testing (api/local)   |"
echo "+-------------------------+"

echo "TODO"
# cd local/
# go test . -v && cd -


echo "+-------------------------+"
echo "|   Testing (api/civo)    |"
echo "+-------------------------+"

echo "TODO"
# cd civo/
# go test . -v && cd -

echo "+-------------------------+"
echo "|   Testing (api/azure)    |"
echo "+-------------------------+"

echo "TODO"
# cd azure/
# go test . -v && cd -

rm -rvf ${HOME}/.ksctl
