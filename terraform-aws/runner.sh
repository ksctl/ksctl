#!/bin/bash
RED="\e[31m"
GREEN="\e[32m"
BLUE="\e[34m"
YELLOW="\e[33m"
ENDCOLOR="\e[0m"

echo -e "\n${YELLOW}You need to manually approve${ENDCOLOR}\n"

echo -e "${GREEN}What is the cluster name${ENDCOLOR}"
read clustername
export TF_VAR_clustername=${clustername}

cd pre-iac/
terraform init
terraform apply


cd ../ksctl
echo -e "${BLUE}database nodes${ENDCOLOR}"
terraform workspace select database; terraform workspace list; terraform apply

echo -e "${BLUE}loadbalancer nodes${ENDCOLOR}"
terraform workspace select loadbalancer; terraform workspace list; terraform apply

echo -e "${BLUE}controlplane nodes${ENDCOLOR}"
terraform workspace select controlplane; terraform workspace list; terraform apply

echo -e "${BLUE}Workerplane nodes${ENDCOLOR}"
terraform workspace select workerplane; terraform workspace list; terraform apply

