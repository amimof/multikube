# Multikube on Kubernetes

Although Multikube is designed to "speak" with multiple Kubernetes cluster, it can run within a Kubernetes cluster itself. As a matter of fact it is the best way of deploying and running Multikube. This is perfect for large environments where Multikube is deployed in a management cluster and routes to multiple application clusters.

This example sets up Multikube with TLS disabled, which isn't recommended but it is the fastest way of getting started. 

## Deploy 

Create a namespace
```
kubectl create namespace multikube
```

Create a secret which holds the kubeconfig. This example will use the one kubectl uses by default on your local computer `~/.kube/config`
```
kubectl create secret generic kubeconfig --from-file ~/.kube/config -n multikube
```

Deploy kubernetes manifest
```
kubectl apply -f https://raw.githubusercontent.com/amimof/multikube/master/deploy/k8s.yaml
```

Port-forward 8080 to test it locally. You can of course create an ingress as well
```
kubectl port-forward deployment/multikube -n multikube 8443:8443
curl -k http://localhost:8443/
no token present in request
```