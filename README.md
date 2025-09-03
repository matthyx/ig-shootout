# ig-shootout

Generate system load on nodes for testing purposes

## Prerequisites
- [ko](https://ko.build/) - A tool for building and deploying Go applications to Kubernetes
- [kubectl](https://kubernetes.io/docs/tasks/tools/) - Command line tool for interacting with Kubernetes clusters
- [kind](https://kind.sigs.k8s.io/) - A tool for running local Kubernetes clusters using Docker container "nodes"

## Deploy to Kubernetes

```bash
KO_DOCKER_REPO=quay.io/kubescape ko apply -B -f config/
```

## Deploy to local kind cluster

```bash
KO_DOCKER_REPO=kind.local ko apply -f config/
```
