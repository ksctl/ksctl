# External MongoDB Storage

Refer : `internal/storage/external/mongodb`

## Data to store and filtering it performs
1. first it gets the cluster data / credentials data based on this filters
    - `cluster_name` (for cluster)
    - `region` (for cluster)
    - `cloud_provider` (for cluster & credentials)
    - `cluster_type` (for cluster)
    - also when the state of the cluster has recieved the stable desired state mark the IsCompleted flag in the specific cloud_provider struct to indicate its done
2. make sure the above things are specified before writing in the storage

## How to use it 
1. you need to call the Init function to get the storage make sure you have the interface type variable as the caller
2. before performing any operations you must call the **Connect()**.
3. for using methods: **Read()**, **Write()**, **Delete()** make sure you have called the **Setup()**
4. for calling **ReadCredentials()**, **WriteCredentials()** you can use it directly just need to specify the cloud provider you want to write
5. for calling **GetOneOrMoreClusters()** you need simply specify the filter
6. for calling **AlreadyCreated()** you just have to specify the func args
7. Don't forget to call the **storage.Kill()** when you want to stop the complte execution. it guarantees that it will wait till all the pending operations on the storage are completed
8. Custom Storage Directory you would need to specify the env var `KSCTL_CUSTOM_DIR_ENABLED` the value must be directory names wit space separated
9. specify the **Required ENV vars**
    - `export MONGODB_USER="<username>"`
    - `export MONGODB_PASSWORD="<password>"`
    - `export MONGODB_DNS="<DNSSeedlist>"`
    also it uses mongodb uri as **mongodb+srv**

## Things to look for
1. make sure when you recieve return data from **Read()**. copy the address value to the storage pointer variable and not the address!
2. When any credentials are written, it will be stored in 
    - Database: `ksctl-{userid}-db`
    - Collection: `{cloud_provider}`
    - Document/Record: `raw bson data` with above specified data and filter fields
3. When any clusterState is written, it gets stored in
    - Database: `ksctl-{userid}-db`
    - Collection: `credentials`
    - Document/Record: `raw bson data` with above specified data and filter fields

4. When you do Switch aka getKubeconfig it fetches the kubeconfig from the point 3 and stores it to `<some_dir>/.ksctl/kubeconfig`

