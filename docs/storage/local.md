# Local Storage

Refer: `internal/storage/local`

## Data to store and filtering it performs
1. first it gets the cluster data / credentials data based on this filters
    - `cluster_name` (for cluster)
    - `region` (for cluster)
    - `cloud_provider` (for cluster & credentials)
    - `cluster_type` (for cluster)
    - also when the state of the cluster has recieved the stable desired state mark the IsCompleted flag in the specific cloud_provider struct to indicate its done
2. make sure the above things are specified before writing in the storage

it is stored something like this
```
 it will use almost the same construct.
 * ClusterInfos => $USER_HOME/.ksctl/state/
	 |-- {cloud_provider}
		|-- {cluster_type} aka (ha, managed)
			|-- "{cluster_name} {region}"
				|-- state.json
 * CredentialInfo => $USER_HOME/.ksctl/credentials/{cloud_provider}.json
```

## How to use it 
1. you need to call the Init function to get the storage make sure you have the interface type variable as the caller
2. before performing any operations you must call the **Connect()**.
3. for using methods: **Read()**, **Write()**, **Delete()** make sure you have called the **Setup()**
4. for calling **ReadCredentials()**, **WriteCredentials()** you can use it directly just need to specify the cloud provider you want to write
5. for calling **GetOneOrMoreClusters()** you need simply specify the filter
6. for calling **AlreadyCreated()** you just have to specify the func args
7. Don't forget to call the **storage.Kill()** when you want to stop the complte execution. it guarantees that it will wait till all the pending operations on the storage are completed
8. Custom Storage Directory you would need to specify the env var `KSCTL_CUSTOM_DIR_ENABLED` the value must be directory names wit space separated
9. it creates the configuration directories on your behalf

## Things to look for
1. make sure when you receive return data from **Read()**. copy the address value to the storage pointer variable and not the address!
2. When any credentials are written, it will be stored in `<some_dir>/.ksctl/credentials/{cloud_provider}.json`
3. When any clusterState is written, it gets stored in `<some_dir>/.ksctl/state/{cloud_provider}/{cluster_type}/{cluster_name} {region}/state.json`
4. When you do Switch aka getKubeconfig it fetches the kubeconfig from the point 3 and stores it to `<some_dir>/.ksctl/kubeconfig`
