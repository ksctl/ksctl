
variable "region" {
  type    = string
  default = "us-east-1"
}

variable "instance_name" {
  type = map(any)
  default = {
    "controlplane" = "control-plane"
    "database"     = "database"
    "loadbalancer" = "loadbalancer"
    "workerplane"  = "worker-plane"
  }
}

variable "control_nodes" {
  default = 3
}

variable "worker_nodes" {
  default = 1
}

variable "number_of_instance" {
  type = map(any)
  default = {
    "controlplane" = 3
    "database"     = 1
    "loadbalancer" = 1
    "workerplane"  = 2
  }
}

variable "clustername" {
  type = string
}

output "public_ips" {
  description = "public dns ips for 'SSH'"
  value = module.instances.aws_instance_public_ip
}
