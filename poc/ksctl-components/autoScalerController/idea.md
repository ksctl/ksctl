# Why are we building the Controller

We want an autoscaler that adds a new virtual machine to the cluster when load increases.

## What are the ways we can do it

- Cluster autoscaler
- Controller

## Why not cluster autoscaler

- works only on managed cluster (except got GCE) its quite limited because it is very much dependent on internal cloud APIs.
- Does not work with Self managed and bootstrapped clusters because cluster autoscaler directly call k8s cloud managed API to intern scale up/down.
- NodePools is used by cluster autoscaler which is not available in many cloud providers like Civo.

## Controller
How to create a controller and what its responsibilities [Refer](https://docs.google.com/document/d/1X7LHlwRENBGIEFDyozRippI-xorBTqOGVjiDeGjYAMI/edit#heading=h.1ye78yl9eiil)

https://github.com/ksctl/ksctl/issues/251

```go
func scaleup(){
    // This contains the logic to scale up the cluster created using ksctl.
}

func scaleDown(){
    // This contains the logic to scale down the cluster created using ksctl.
}
```

### Controller

It decides whether the cluster needs to be Upscaled or downscaled

It is also responsible for calling the functionality for it.

## How do we know when to scale up

1. We check **pod pressure** via the Spec.Status.Condition[\...\]
> **Note**: Pod pressure is calculated by checking the number of pending pods in the cluster and seeing if the current nodes in the cluster can schedule it. If not, new nodes need to be added to the cluster.

## How do we know when to scale down
Factors we are lokking forward:
1. **Underutilization Nodes `PRIMARY`**: resource usage is sum(resource_requests) / node_allocatable (not yet decided how!); It has nothing to do with "real" utilization
> **Note**: if it's underutilized over a certain period. This involves checking if the node's CPU, memory, and other resources are not being effectively used by the pods scheduled on it.
2. **Health Status**: Nodes experiencing issues or deemed unhealthy based on specific criteria might be terminated to maintain the overall health of the cluster
3. **Cost Efficiency**: for it to work we need to add node labels so that we can determine which node is of higher specs or we can also do use the avail mem and cpu resources.  (for instance, larger nodes that are underutilized) might be targeted for termination to reduce costs.

### How Controller talks to the Ksctl Agent

We can use event based mechanism, where the agent keeps watching for resources that hold the desired state of the cluster.

The desired state can be:
- Number of Nodes
- or any other future needs.

When the resource changes, ksctl agent is responsible for reconciling the desired and current state of the cluster.

If the cluster needs to be scaled up, ksctl agent will do it.

### The process of scaling up.

- Managed cluster (Cluster autoscaler).
- Self Managed Cluster

#### Self Managed Cluster

- Cloud resource will be created (worker plane VMs).
- Use KubeAdm or K3S to bootstrap the above VM.
> **Note**
> We need the latest cluster token to add a new VM to the existing cluster.
> we can fetch the latest token by ssh into controlplane-0 and then use to join the Worker Plane.