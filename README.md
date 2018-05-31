# multikube

Define a set of clusters and create a cluster group
```Go
group := multikube.NewGroup("dev").AddCluster(&multikube.Config{
  Name: "minikube",
  Hostname: "https://192.168.99.100:8443",
  Cert: "/Users/amir/.minikube/client.crt",
  Key: "/Users/amir/.minikube/client.key",
  CA: "/Users/amir/.minikube/ca.crt",
}, &multikube.Config{
  Name: "prod-cluster-1",
  Hostname: "https://192.168.99.100:8443",
  Cert: "/Users/amir/.minikube/client.crt",
  Key: "/Users/amir/.minikube/client.key",
  CA: "/Users/amir/.minikube/ca.crt",
})
```

Sync the cache for all clusters
```Go
err := group.SyncAll()
if err != nil {
  panic(err)
}
```