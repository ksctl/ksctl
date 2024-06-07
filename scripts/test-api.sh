#!/bin/sh

EXEC=$1

cd ../pkg/ || exit 1

echo "-----------------------------------"
echo "|   Testing (pkg/helpers)"
echo "-----------------------------------"

cd helpers/
GOTEST_PALETTE="red,yellow,green" $EXEC -fuzz=Fuzz -fuzztime 10s -v cloud_test.go fields.go
GOTEST_PALETTE="red,yellow,green" $EXEC -fuzz=Fuzz -fuzztime 10s -v cni_test.go fields.go
GOTEST_PALETTE="red,yellow,green" $EXEC -fuzz=Fuzz -fuzztime 10s -v name_test.go fields.go
GOTEST_PALETTE="red,yellow,green" $EXEC -fuzz=Fuzz -fuzztime 10s -v storage_test.go fields.go
GOTEST_PALETTE="red,yellow,green" $EXEC -fuzz=Fuzz -fuzztime 10s -v distro_test.go fields.go
GOTEST_PALETTE="red,yellow,green" $EXEC -fuzz=Fuzz -fuzztime 10s -v role_test.go fields.go
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -

echo "-----------------------------------"
echo "|   Testing (pkg/logger)"
echo "-----------------------------------"

cd logger/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v -timeout 10s && cd -

cd ../internal

echo "--------------------------------------------"
echo "|   Testing (internal/k8sdistros/k3s)"
echo "--------------------------------------------"

cd k8sdistros/k3s/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -

echo "--------------------------------------------"
echo "|   Testing (internal/cloudproviders/local)"
echo "--------------------------------------------"

cd cloudproviders/local/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -


echo "--------------------------------------------"
echo "|   Testing (internal/cloudproviders/civo)"
echo "--------------------------------------------"

cd cloudproviders/civo/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -

echo "--------------------------------------------"
echo "|   Testing (internal/cloudproviders/azure)"
echo "--------------------------------------------"

cd cloudproviders/azure/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -

echo "--------------------------------------------"
echo "|   Testing (internal/cloudproviders/aws)"
echo "--------------------------------------------"

cd cloudproviders/aws/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -

echo "--------------------------------------------"
echo "|   Testing (internal/storage/local)"
echo "--------------------------------------------"

cd storage/local/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -

echo "--------------------------------------------"
echo "|   Testing (internal/storage/kubernetes)"
echo "--------------------------------------------"

cd storage/kubernetes/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -


echo "-------------------------------------------------"
echo "|   Testing (internal/storage/external/mongodb)"
echo "-------------------------------------------------"

cd storage/external/mongodb/
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -

cd ..


echo "-------------------------------------------------"
echo "|   Testing (ksctl-components/agent)"
echo "-------------------------------------------------"

cd ksctl-components/agent
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -

echo "-------------------------------------------------"
echo "|   Testing (ksctl-components/stateimport)"
echo "-------------------------------------------------"

cd ksctl-components/stateimport
GOTEST_PALETTE="red,yellow,green" $EXEC . -v && cd -

