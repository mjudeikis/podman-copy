# Virtual Kubelet Podman Provider

This is a Virtual Kubelet Provider implementation that manages containers in podman environment

## Purpose

The whole point of the Virtual Kubelet project is to provide an interface for container runtimes that don't conform to the standard node-based model. The [Kubelet](https://github.com/kubernetes/kubernetes/tree/master/pkg/kubelet) codebase is the comprehensive standard CRI node agent and this Provider is not attempting to recreate that..

## Dependencies

Podman TBC

## Configuring

* Copy `/etc/kubernetes/admin.conf` from your master node and place it somewhere local to Virtual Kubelet
* Find a `client.crt` and `client.key` that will allow you to authenticate with the API server and copy them somewhere local

## Running

Start podman
```cli
TBC
```
Create a script that will set up the environment and run Virtual Kubelet with the correct provider
```
#!/bin/bash
export VKUBELET_POD_IP=<IP of the Linux node>
export APISERVER_CERT_LOCATION="/etc/virtual-kubelet/client.crt"
export APISERVER_KEY_LOCATION="/etc/virtual-kubelet/client.key"
export KUBELET_PORT="10250"
cd bin
./virtual-kubelet --provider podman --nodename podman --provider-config ./hack/podman-cfg.json --kubeconfig admin.conf 
```


## Limitations


## Pre-reqa

```
yum install podman
systemctl enable --now io.podman.socket
systemctl status io.podman.socket
```

I have enabled my local dev user to access socket for development:
```
[Socket]
ListenStream=/run/podman/io.podman
SocketMode=0770
SocketGroup=root
SocketUser=$USER
```

`vi /etc/tmpfiles.d/podman.conf`:
```
d /run/podman 0770 root $USER
```
