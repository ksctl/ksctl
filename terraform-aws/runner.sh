#!/bin/bash

echo "What is the cluster name"
read clustername
export TF_VAR_clustername=${clustername}
cd ksctl
terraform fmt
terraform workspace select database
terraform workspace list
terraform init
terraform plan

terraform workspace select loadbalancer
terraform workspace list
terraform init
terraform plan

echo "controlplane nodes"

terraform workspace select controlplane
terraform workspace list
terraform init
terraform plan

echo "Workerplane nodes"

terraform workspace select workerplane
terraform workspace list
terraform init
terraform plan

