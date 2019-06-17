## Installing Che for Codewind on Kubernetes

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