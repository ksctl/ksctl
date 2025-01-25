terraform {
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "1.47.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.6.2"
    }
  }
  # https://developer.hashicorp.com/terraform/language/settings/backends/s3
  backend "s3" {
    bucket  = "ksctl-e2e-tf-0115"
    region  = "eu-north-1"
    key     = "ksctl/e2e/terraform.tfstate"
    encrypt = true
  }
}

provider "hcloud" {
  token = var.hcloud_token
}

provider "random" {
}

resource "hcloud_firewall" "sshfirewall" {
  name = "${var.resource_name}-fw"
  rule {
    direction  = var.ssh_inbound.direction
    protocol   = var.ssh_inbound.protocol
    source_ips = var.ssh_inbound.source_ips
    port       = var.ssh_inbound.port
  }
}

resource "hcloud_ssh_key" "ssh_keypair" {
  name       = "${var.resource_name}-ssh"
  public_key = file(var.ssh_pub_loc)
}

resource "random_string" "random_id" {
  count   = var.no_servers
  length  = 8
  special = false
  upper   = false
  lower   = true
  numeric = true
}

resource "hcloud_server" "server" {
  name         = "${var.resource_name}-${random_string.random_id[count.index].result}"
  server_type  = var.machine_type
  image        = var.os_version
  firewall_ids = [hcloud_firewall.sshfirewall.id]
  location     = var.vm_location
  ssh_keys     = [hcloud_ssh_key.ssh_keypair.id]
  count        = var.no_servers
  user_data    = <<EOF
    #!/bin/bash
    sudo apt update -y
  EOF

  public_net {
    #ipv6_enabled = true
    ipv4_enabled = true
  }
}

output "ip_address" {
  value     = hcloud_server.server.*.ipv4_address
  sensitive = true
}

# output "ssh_command" {
#   value = "ssh -i ${var.ssh_pvt_loc} root@${hcloud_server.server.ipv6_address}"
# }
