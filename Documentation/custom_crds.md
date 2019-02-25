## Operate Custom Resources

In order for the GitOps operator to operate on other CRDs (Custom Resource Definitions) you need to give those as inputs to the operator so that the API objects are known to the client.

Apply the configmap and delete the gitops operator pod to make load the changes
```
kubectl apply -f deploy/configmap.yaml
kubectl delete po 
```

Deploy a GitOps CR that orchestrates a prometheus instance through Prometheus-operator. GitOps operator is able to create the custom resource prometheus when deployed with configmap
```
kubectl apply -f deploy/crd/prom_ops.yaml
```
- Check that prometheus-operator and a prometheus instance gets created.

