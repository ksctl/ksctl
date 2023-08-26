# Roadmap

## Current Status on Supported Providers

export const Highlight = ({children, color}) => (
  <span
    style={{
      backgroundColor: color,
      borderRadius: '2px',
      color: '#fff',
      padding: '0.2rem',
    }}>
    {children}
  </span>
);

<Highlight color="green">Done</Highlight> <Highlight color="red">Not Started</Highlight> <Highlight color="black">No Plans</Highlight> <Highlight color="blue">Backlog</Highlight>


```mermaid
flowchart LR;
  classDef green color:#022e1f,fill:#00f500;
  classDef red color:#022e1f,fill:#f11111;
  classDef white color:#022e1f,fill:#fff;
  classDef black color:#fff,fill:#000;
  classDef blue color:#fff,fill:#00f;

  XX[ksctl]--CloudFactory-->web{Cloud Providers};
  XX[ksctl]--DistroFactory-->web2{Distributions};
  XX[ksctl]--StorageFactory-->web3{State Storage};

  web--Civo-->civo{Types};
  civo:::green--managed-->civom[Create & Delete]:::green;
  civo--HA-->civoha[Create & Delete]:::green;

  web--Local-Kind-->local{Types};
  local:::green--managed-->localm[Create & Delete]:::green;
  local--HA-->localha[Create & Delete]:::black;

  web--AWS-->aws{Types};
  aws:::blue--managed-->awsm[Create & Delete]:::red;
  aws--HA-->awsha[Create & Delete]:::red;

  web--Azure-->az{Types};
  az:::green--managed-->azsm[Create & Delete]:::green;
  az--HA-->azha[Create & Delete]:::green;

  web2--K3S-->k3s{Types};
  k3s:::green--HA-->k3ha[Create & Delete]:::green;

  web2--Kubeadm-->kubeadm{Types};
  kubeadm:::blue--HA-->kubeadmha[Create & Delete]:::red;

  web3--Local-Store-->slocal{Local}:::green;
  web3--Remote-Store-->rlocal{Remote}:::red;

```



## Future Plans

:::note
The below list is for most probable goals
:::

0. Add HA external Datastore (use of etcd instead of mysql)
1. add distributions of binaries via the package managers
2. AWS Support
3. GCP support
4. improve context switch
5. Cloud Controller support for more control
6. Create RestAPI for using ksctl without CLI
7. Create server running ksctl which can reconsile state automatically


### CheckList (not exaustive)
- [ ] add Web client
- [ ] GCP
- [ ] AWS
- [x] additional kubernetes application support
- [ ] all other cloud providers
- [x] improve the High avilability cluster architecture
- [x] improve logging in local

