each VM has
1. of its own
  - itself
  - disk
  - Public IP / None
  - network interface

2. reuse resources
  - Resourecegroup
  - Virtualnetwork which is bound to that resource group
  - SSH key pair
  - network security group
  - use the default subnet

use the same nsg for all controlplane, workerplane and database, loadbalancer
