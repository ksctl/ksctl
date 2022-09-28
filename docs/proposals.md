# How all things should work

* must use VMs for Cloud Providers as well as Local Machine(as much as possible) 
	> EXCEPTION: docker containers due to resource limits

* CLI should be able to `create-cluster`, `view-cluster`, `delete-cluster`

> **NOTE**

> Suggestions are most Welcome

# WorkFlow

1. CLI will get the command from the user
2. Some processing program will check request's validation
3. Then CLI will call using API to create job which in-turn will allocate the neccessary resources and configure them
4. Allocation will most likely be **Each Node in seperate VM**

> **NOTE**

> Suggestions are most Welcome



# Initial Goal

* **Single Node cluster** Local system

# Desired Goal

* Run the HA Kubernetes Cluster with multiple VMs in cloud platforms
* Run HA Kubernetes cluster on Local Machine **if resource permit**

Hope you have Great time Contributing :heart:
