resource "aws_instance" "instance"{
  ami = var.ami
  key_name = "${var.cluster_name}"
  instance_type = var.name == "database" ? var.instance_type[2] : var.name == "controlplane" ? var.instance_type[1] : var.instance_type[0]
  availability_zone = var.region_az

  count = var.no_of_vms
  security_groups = ["${var.cluster_name}-sg-${var.name}"]
  tags = {
    "Name" = "ha-ksctl-${var.cluster_name}-${var.name}-${count.index+1}"
  }
  user_data = var.name == "database" ? file("../modules/scripts/database.sh") : var.name == "loadbalancer" ? file("../modules/scripts/loadbalancer.sh") : <<-EOF
    #!/bin/bash
    sudo apt update -y
  EOF
}

data "local_file" "k3s_password" {
  filename = "password.txt"
}

output "aws_instance_public_ip" {
  description = "instance public IP"
  value = aws_instance.instance.*.public_dns
}
