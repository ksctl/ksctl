# Export Credential

> should look like a endpoint for the ksctl agent to reuse it to establish its connection with the external db

## for MongoDB

```
mongodb+srv://${username}:${password}@${domain}
```

```
mongodb://${username}:${password}@${domain}:${port}
```

## For redis

```
redis://${password}@${host}:${port}
```

what we can do is we can attach a method or a simple function call to get a endpoint string
```go
func ExportEndpoint() string
```

this can be used for the setup of the state to the cluster (ksctl or non-Ksctl) cluster
with external store


