resource "random_password" "k3s_db_password" {
  length = 20
}

variable "clustername" {
  type = string
}

variable "region" {
  type    = string
  default = "us-east-1"
}

resource "local_file" "password" {
  content = random_password.k3s_db_password.result
  filename = "../ksctl/password.txt"
  file_permission = "0440"
}

resource "tls_private_key" "ssh" {
  algorithm = "RSA"
  rsa_bits  = "4096"
}

resource "local_file" "private_key" {
  content         = tls_private_key.ssh.private_key_openssh
  filename        = "../ksctl/${var.clustername}.pem"
  file_permission = "0400"
}

resource "aws_key_pair" "key_pair" {
  depends_on = [
    local_file.private_key,
  ]
  key_name   = var.clustername
  public_key = tls_private_key.ssh.public_key_openssh
}
