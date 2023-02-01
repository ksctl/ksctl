provider "aws" {
  region = var.region
}

provider "local" {}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.52.0"
    }
  }
}

provider "random" {}