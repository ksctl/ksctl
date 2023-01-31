
variable instance_type {
  type = list(string)
  default = ["t2.micro", "t2.medium", "t2.large"]
}

variable ami {
  type = string
  default = "ami-00874d747dde814fa"
}

variable name {
  type = string
}

variable region_az {
  type = string
  default = "us-east-1a"
}
variable "keypair"{
  type = string
}

variable "cluster_name" {
  type = string
}

variable no_of_vms {
  type = number
}