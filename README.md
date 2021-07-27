# kubedeploy
Simple blue/green deployment kubectl plugin

`kube-deploy` helps you to implement blue/green deployment in your k8s cluster:

### Installation

#### Manual


### How it works


### Usage

### Plugin configurations

### Developing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to contribute to this project.

You can develop locally with
[Minikube](https://kubernetes.io/docs/setup/minikube/).

On Linux, the `kvm2` driver provides better performance than the default
`virtualbox` driver, but either will work:

```
minikube start --vm-driver=kvm2
```

`minikube start` will configure your `kubeconfig` for your local Minikube
cluster and set the current context to be for Minikube.

### License

Kubectl Deployments plugin is [MIT licensed](LICENSE).

### Authors

* Afshin Paydar