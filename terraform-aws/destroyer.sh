#!/bin/bash

echo "What is the cluster name"
read clustername
export TF_VAR_clustername=${clustername}

cd ksctl

echo "database nodes"

terraform workspace select database
terraform workspace list
terraform destroy

echo "loadbalancer nodes"
terraform workspace select loadbalancer
terraform workspace list
terraform destroy

echo "controlplane nodes"

terraform workspace select controlplane
terraform workspace list
terraform destroy

echo "Workerplane nodes"

terraform workspace select workerplane
terraform workspace list
terraform destroy


cd ../pre-iac
terraform destroy --auto-approve
