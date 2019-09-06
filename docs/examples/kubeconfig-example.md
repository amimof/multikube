# kubeconfig

Multikube uses the well-known [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) to communicate with upstream apiservers. 

A request from a client with the URL path `/k8s-dev-cluster/api/v1/namespaces` will tell Multikube to find a context in its configured kubeconfig named 'k8s-dev-cluster'. 

## Create a kubeconfig for Multikube

Create a cluster
```
kubeconfig config set-cluster k8s-dev-cluster \
  --server=https://k8s-dev.domain.com:8443 \
  --certificate-authority=ca.crt \
  --embed-certs=true \
  --kubeconfig=/etc/multikube/kubeconfig
```

Create the credentials. This example uses client certificate pairs
```
kubectl config set-credentials k8s-dev-cluster \
  --client-certificate=client.crt \
  --client-key=client.key \
  --embed-certs=true \
  --kubeconfig=/etc/multikube/kubeconfig
```

Create the context 
```
kubectl config set-context k8s-dev-cluster \
  --cluster=k8s-dev-cluster \
  --user=k8s-dev-cluster \
  --kubeconfig=/etc/multikube/kubeconfig
```

## Run Multikube

We can now use the kubeconfig we created with Multikube. Clients may target request to the cluster named `k8s-dev-cluster` either by using URL path or an HTTP header. 

```
./multikube 
```