---
sidebar_position: 1
---

# ksctl Intro

It's a CLI tool which can manage Kubernetes cluster running on different environment(cloud platforms)

**ksctl takes less than 15 minutes**.

## Getting Started

Lets start to understand things better

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

  XX[CLI]:::white--providers-->web{API};
  web:::white--CIVO-->civo{Types};
  civo:::green--managed-->civom[Create & Delete]:::green;
  civo--HA-->civoha[Create & Delete]:::green;

  web--LOCAL-->local{Types};
  local:::green--managed-->localm[Create & Delete]:::green;
  local--HA-->localha[Create & Delete]:::black;

  web--AWS-->aws{Types};
  aws:::blue--managed-->awsm[Create & Delete]:::red;
  aws--HA-->awsha[Create & Delete]:::red;

  web--AZURE-->az{Types};
  az:::green--managed-->azsm[Create & Delete]:::green;
  az--HA-->azha[Create & Delete]:::green;

```

## Current Features

Having core features of `get`, `create`, `delete`, `add nodes`, `switch`, under specific providers

- Local
    - kind cluster with specific version
- Civo
    - have support for the managed and High Available clusters(_Custom_)
- Azure
    - have support for the managed and High Available clusters(_Custom_)

## Future Plans
- add Web client
- GCP
- AWS
- additional kubernetes application support
- all other cloud providers
- improve the High avilability cluster architecture
- improve logging in local


## Issues and current work
- work on improving the testing
- look for labels `#priority/essential`, `#priority/should_have` and `#kind/bug`

## Current Releases

- [ ] 1.0
- [x] 1.0-rc2
- [x] 1.0-rc1
- [x] ...
