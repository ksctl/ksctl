variable "hcloud_token" {
  sensitive = true
  type      = string
}

variable "machine_type" {
  default     = "cx22"
  description = "machine type for runner"
  type        = string
}

variable "resource_name" {
  default     = "e2e-runner"
  description = "e2e runner resource name"
  type        = string
}

variable "os_version" {
  default     = "ubuntu-24.04"
  description = "os version"
  type        = string
}

variable "no_servers" {
  default = 3
  type    = number
}

variable "vm_location" {
  default     = "nbg1"
  description = "region for vm"
  type        = string
}

variable "ssh_pub_loc" {
  default     = "~/.ssh/demo/id_rsa.pub"
  description = "ssh pub loc"
  type        = string
}

variable "ssh_pvt_loc" {
  default     = "~/.ssh/demo/id_rsa"
  description = "ssh pvt loc"
  type        = string
}

variable "ssh_inbound" {
  type = object({
    direction  = string
    protocol   = string
    port       = string
    source_ips = list(string)
  })
  default = {
    direction  = "in"
    protocol   = "tcp"
    port       = "22"
    source_ips = ["0.0.0.0/0", "::/0"]
  }
}
