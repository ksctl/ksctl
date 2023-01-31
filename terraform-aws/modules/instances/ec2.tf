resource "aws_instance" "instance"{
  ami = var.ami
  key_name = "${var.cluster_name}"
  instance_type = var.instance_type[0]
  availability_zone = var.region_az

  count = var.no_of_vms
  security_groups = ["${var.cluster_name}-sg"]
  tags = {
    "Name" = "${var.name}-${count.index+1}"
  }
  user_data = <<-EOF
    #!/bin/bash
    sudo apt update -y
    sudo apt install vim neovim -y
  EOF
}

resource "tls_private_key" "ssh" {
  algorithm = "RSA"
  rsa_bits  = "4096"
}

resource "local_file" "private_key" {
  content         = tls_private_key.ssh.private_key_openssh
  filename        = "${var.keypair}.pem"
  file_permission = "0400"
}

resource "aws_key_pair" "key_pair" {
  depends_on = [
    local_file.private_key
  ]
  key_name = "${var.cluster_name}"
  public_key = tls_private_key.ssh.public_key_openssh
}

output "aws_instance_public_key" {
  description = "instance public IP"
  value = aws_instance.instance.*.public_dns
}
