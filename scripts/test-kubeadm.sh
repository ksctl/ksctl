#!/bin/sh

cd ../internal/ || exit 1

echo "--------------------------------------------"
echo "|   Testing (internal/k8sdistros/kubeadm)"
echo "--------------------------------------------"

cd k8sdistros/kubeadm/
go test . -v && cd -

