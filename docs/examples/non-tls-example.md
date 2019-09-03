## Non-TLS example (not recommended)
This method of serving multikube is not recommended for obvious reasons. Additionally, this method prevents kubectl compatibility due to how kubectl discards the `token` field in kubectl when the cluster has a http scheme over https.

```
multikube --scheme=http 
```