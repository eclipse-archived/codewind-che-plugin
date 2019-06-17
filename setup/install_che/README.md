## Installing Che for Codewind on Kubernetes

The `./install.sh` script provides a simple way to install a Codewind-ready version of Eclipse Che on Kubernetes. It installs the latest version of Eclipse Che, with a special Cluster Role set for Eclipse Codewind. 

### Prerequisites

1. Kubernetes cluster with ingress installed
2. Active kubectl context to the cluster

### Install

1. Determine your ingress domain. It should be of the form <IP>.nip.io.
    - If you're running on IBM Cloud Private, this will be the public IP address of your proxy node.
    - On other Kubernetes, use your master node IP address
2. Configure kubectl for your cluster
    - This will depend on your cluster
    - On OpenShift use `oc login`, on IBM Cloud Private, use `cloudctl login`.
3. Run `./install.sh` to deploy Eclipse Che
    - `./install.sh -h` will show the available CLI options
    - By default, it runs a `helm install` of Eclipse Che, but you can configure the install method used with the `-m` flag.