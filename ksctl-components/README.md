### Work notes
- first work on adding automation to the deployment
    of ksctl agent and the storage importer
- Add Application stack for applications

---
# Ksctl-Operators

Domain: `ksctl.com`

> [!WARNING]
> Currently we havn't came to a conclusion on how to deploy the controller
> in the automated via via the kubernetes client


## Ksctl agent

## Controllers & CRDs

### Application Stacks and Components

#### Application.Stacks()
apiVersion: `storage.ksctl.com/v1alpha1`
kind: `ImportState`

component or stacks??

how should we format it 
thing is we want to controller how all the ksctl management toolsa re created 

- one option is to use the application name to install
  - here the override options will be available for more fine grain options (**My guess is it will be used in production stack setting**)
- another option is to use give users more control
  - not quite sure how it fit well
  - also we dont want to compete with tools already present like argocd and gitops

> use these notes in upcomming calls to clarify them

```yaml
spec:
  components:
  - name: !str ""
    override: !map[string]any
```

### Storage Importer & Exporter

> [!NOTE]
> for now we are going to use this just for exporting the state files
> (given) the creation of the cluster took place from host local machine
> (constrains) it will not import when the storage falls under **_external storage compatability requirements_**

#### Importer
it will [Watch](#storageimport)
will create a ksctl agent rpc client to send the docuemnt to import in the kubernetes cluster

### Loadbalancer Provisioning

> [!NOTE]
> Work in Progress

### Cluster autoscaler

> [!NOTE]
> Work in Progress

### Storage Exporter and Importer

##### How to install it?

> go to the root of the project

```bash
docker build --file build/agent/Dockerfile --tag ghcr.io/ksctl/ksctl-agent:latest .
sudo docker push ghcr.io/ksctl/ksctl-agent:latest

make -f Makefile.controllers docker-build
sudo docker push ghcr.io/ksctl/ksctl:controller-storage-latest
```

> to install crds and controller

```bash
cd ksctl-components/operators
make install
cd ../..

make -f Makefile.controllers deploy
```

> to deploy the ksctl agent

```bash
cd ksctl-components/agent
kubectl apply -f example-deploy-agent.yml
```

> for logs

```bash
k logs -n operators-system deployment/operators-controller-manager -f
k logs -nksctl deployments/ksctl-agent -f
```


> the trigger is custom resource

```bash
cd ksctl-components/operators
k apply -k config/samples/
```

##### How to Uninstall it?

```bash
# on the root of the project
make -f Makefile.controllers deploy

cd ksctl-components/agent
kubectl delete -f example-deploy-agent.yml
```


#### Storage.Import()
apiVersion: `storage.ksctl.com/v1alpha1`
kind: `ImportState`

```bash
kubectl proxy

curl localhost:8001/apis/storage.ksctl.com/v1alpha1/importstates | jq -r .
```

##### CRD install
```bash
kubectl apply -f ksctl-components/operators/config/crd/bases/storage.ksctl.com_importstates.yaml
```

#### Storage.Export()

> [!NOTE]
> Work in Progress
