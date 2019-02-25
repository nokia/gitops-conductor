## GitOps operator

GitOps conductor is a Kubernetes operator using the operator pattern for managing resources on a Kubernetes cluster. It builds on the concept introduced by [WeaveWorks](https://www.weave.works/blog/gitops-operations-by-pull-request). The GitOps conductor philosophy is to build on the Unix philosohy and does not try to solve your whole CI/CD pipeline. GitOps conductor only takes care of ensuring the thing you want on your cluster is there.

The conductor works with read only permissions to the git repositories applied on the cluster. This philosophy is kept to ensure that a hacked cluster would not give a hacker possibility to hijack other clusters by pushing code to your deployment repo.

---

## Get Started

Deploy the GitOps CRD and the operator modify the RBAC rules if you want to restrict access

```
kubectl apply -f deploy/crds/ops_v1alpha1_gitops_crd.yaml 
kubectl apply -f deploy/cluster-role.yaml
kubectl apply -f deploy/role_binding.yaml
kubectl apply -f deploy/role.yaml
kubectl apply -f deploy/service_account.yaml
kubectl apply -f deploy/operator.yaml
```

Once the GitOps conductor is running you can deploy a CR instance to get the operator into work.

A simple example that deploys (and keeps) a busybox pod in your cluster
```
kubectl apply -f deploy/crds/ops_v1alpha1_gitops_cr.yaml
```
