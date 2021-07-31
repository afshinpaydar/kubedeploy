# kubedeploy
Simple blue/green deployment kubectl plugin

`kube-deploy` plugin helps you to manage deployments in your k8s cluster:

### Blue/Green deployment
`kubectl-deploy bluegreen` expect two Deployments and one Service, that points to one of those in the active k8s cluster
the name of Deployments and Service doesn’t matter and could be anything,
and also how the Service exposed to outside of Kubernetes cluster.

![Blue/Green](img/blue-green.png?raw=true "Blue/Green Deployment")

### Installation
```
$ make
$ make install
$ kubectl deploy version
$ docker run -it kubectl-deploy:latest deploy version
```
#### Manual (Linux)
```sh
$ curl -sS  https://github.com/afshinpaydar/kubedeploy/releases/latest/kubectl-deploy-x86-64-linux -o kubectl-deploy
$ sudo mv kubectl-deploy /usr/local/bin/
$ chmod +x kubectl-deploy
$ kubectl deploy version
```

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