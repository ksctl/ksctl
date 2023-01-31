module "instances" {
  source = "../modules/instances"
  name = lookup(var.instance_name, terraform.workspace)
  no_of_vms = lookup(var.number_of_instance, terraform.workspace)
  keypair = var.clustername
  cluster_name = var.clustername
}