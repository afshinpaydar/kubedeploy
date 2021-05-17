# kubedeploy
Simple blue/green deployment plugin

`kube-deploy` helps you to implement blue/green deployment in your k8s cluster:

### Installation

#### Manual

Since `kube-deploy` is written in Bash, you should be able to install it to any POSIX environment that has Bash installed.


- Download the `kube-deploy` script.
- Either:
  - save it to somewhere in your `PATH`,
  - or save it to a directory, then create symlinks to `kube-deploy` from
    somewhere in your `PATH`, like `/usr/local/bin`
- Make `kube-deploy` executable (`chmod +x ...`)

```bash
$ git clone git@github.com:afshinpaydar/kubedeploy.git
$ cp kubedeploy/bin/kubectl-deploy /usr/local/bin/
$ chmod +x /usr/local/bin/kubectl-deploy
```

### Usage

```bash
$ kubectl deploy -h
Usage: kubectl-deploy [-s <server>] [-c <CA path>] [-n <namespace>] [-T <token>] [-t <timeout>] [-l <tag>] [-d <docker_repo>] <service> <image>
Arguments:
service REQUIRED: The name of the service the script should trigger the Blue/Green deployment
image REQUIRED: Name of Docker image
-n OPTIONAL: the namespace scope for this CLI request, default is the 'default' namespace
-l OPTIONAL: The new docker tag to deployment, defaults is the 'latest' tag
-t OPTIONAL: How long to wait for the deployment to be available, defaults to 120 seconds, must be greater than 60
-d OPTIONAL: Name of Docker repository, default is "docker.io/kubedeploy"
-s OPTIONAL: The address and port of the Kubernetes API server
-c OPTIONAL: Path to a cert file for the certificate authority
-T OPTIONAL: Token for authentication to the K8S API server

```
