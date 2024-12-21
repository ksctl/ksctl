## How to run it?

> [!TODO]
> Use terraform workspace

## Setup

```bash
export AWS_ACCESS_KEY_ID=""
export AWS_SECRET_ACCESS_KEY=""
export TF_VAR_hcloud_token=""
export TF_VAR_ssh_pub_loc=""
```

## Apply | Destroy

```bash
terraform init
```

```bash
terraform workspace list
terraform workspace select dev
terraform workspace new dev
terraform workspace new prod
```

```bash
terraform plan
```

```bash
terraform apply
terraform destroy
```

```bash
terraform output -json ip_address | jq -r .
```
